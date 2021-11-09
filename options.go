package closer

import (
	"io"
	"os"
)

func WithClosers(closers ...io.Closer) Option {
	return func(c *Closer) {
		c.closers = append(c.closers, closers...)
	}
}

func WithCloseFuncs(funcs ...func()) Option {
	return func(c *Closer) {
		c.closeFuncs = append(c.closeFuncs, funcs...)
	}
}

func WithSignals(sigs ...os.Signal) Option {
	return func(c *Closer) {

	}
}
