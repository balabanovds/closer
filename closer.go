package closer

import (
	"context"
	"io"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Closer contains all service closers to handle 'em gently on signal received from OS
type Closer struct {
	sync.Mutex
	once       sync.Once
	closers    []io.Closer
	closeFuncs []func()
	sigCh      chan os.Signal
	sigs       []os.Signal
	timeout    time.Duration
	done       chan struct{}
}

type Option func(*Closer)

// New ...
func New(timeout time.Duration, opts ...Option) *Closer {
	c := Closer{
		sigs:    []os.Signal{syscall.SIGINT, syscall.SIGTERM},
		sigCh:   make(chan os.Signal, 1),
		timeout: timeout,
		done:    make(chan struct{}),
	}

	for _, opt := range opts {
		opt(&c)
	}

	go func() {
		signal.Notify(c.sigCh, c.sigs...)

		sig := <-c.sigCh
		log.Printf("received syscall signal: %s", sig.String())
		c.drop()
	}()

	return &c
}

// AddCloser any io.Closer
func (c *Closer) AddCloser(cl io.Closer) {
	c.Lock()
	defer c.Unlock()
	c.closers = append(c.closers, cl)
}

// AddFunc any func() that will be run on exit
func (c *Closer) AddFunc(f func()) {
	c.Lock()
	defer c.Unlock()
	c.closeFuncs = append(c.closeFuncs, f)
}

// Close force close all underlying Closers
func (c *Closer) Close() {
	c.drop()
}

func (c *Closer) Wait() {
	<-c.done
}

func (c *Closer) drop() {
	c.once.Do(func() {
		c.Lock()
		defer c.Unlock()

		var wg sync.WaitGroup

		ctx, cancel := context.WithTimeout(context.Background(), c.timeout)

		go func() {
			wg.Wait()
			cancel()
		}()

		for _, cl := range c.closers {
			wg.Add(1)
			go func(cl io.Closer) {
				defer wg.Done()
				err := cl.Close()
				if err != nil {
					log.Printf("failed to close: %s", err)
				}
			}(cl)
		}

		for _, cf := range c.closeFuncs {
			wg.Add(1)
			go func(f func()) {
				defer wg.Done()
				f()
			}(cf)
		}

		log.Printf("waiting %s before terminate or end up earlier if funcs ready", c.timeout.String())
		<-ctx.Done()

		c.done <- struct{}{}
	})
}
