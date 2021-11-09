package closer

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"
)

//Closer contains all service closers to handle 'em gently on signal received from OS
type Closer struct {
	sync.Mutex
	closers    []io.Closer
	closeFuncs []func()
	sigCh      chan os.Signal
	sigs       []os.Signal
	timeout    time.Duration
	log        *zap.Logger
}

type Option func(*Closer)

//New ...
func New(log *zap.Logger, timeout time.Duration, opts ...Option) *Closer {
	c := Closer{
		sigs:    []os.Signal{syscall.SIGINT, syscall.SIGTERM},
		sigCh:   make(chan os.Signal, 1),
		timeout: timeout,
		log:     log,
	}

	for _, opt := range opts {
		opt(&c)
	}

	go func() {
		signal.Notify(c.sigCh, c.sigs...)

		sig := <-c.sigCh
		c.log.Info("received syscall signal", zap.String("sig", sig.String()))
		c.drop()
	}()

	return &c
}

//AddCloser any io.Closer
func (c *Closer) AddCloser(cl io.Closer) {
	c.Lock()
	defer c.Unlock()
	c.closers = append(c.closers, cl)
}

//AddFunc any func() that will be run on exit
func (c *Closer) AddFunc(f func()) {
	c.Lock()
	defer c.Unlock()
	c.closeFuncs = append(c.closeFuncs, f)
}

//Close force close all underlying Closers
func (c *Closer) Close() {
	c.drop()
}

func (c *Closer) drop() {
	c.Lock()
	defer c.Unlock()

	for _, cl := range c.closers {
		go func(cl io.Closer) {
			err := cl.Close()
			if err != nil {
				c.log.Error("failed to close", zap.Error(err))
			}
		}(cl)
	}

	for _, cf := range c.closeFuncs {
		go cf()
	}

	c.log.Info(fmt.Sprintf("waiting %s before terminate", c.timeout.String()))
	time.Sleep(c.timeout)

	os.Exit(1)
}
