package closer

import (
	"io"
	"os"
)

// WithClosers add any io.Closer to internal state
func WithClosers(closers ...io.Closer) Option {
	return func(c *Closer) {
		c.closers = append(c.closers, closers...)
	}
}

// WithCloseFuncs add any func to internal state
func WithCloseFuncs(funcs ...func()) Option {
	return func(c *Closer) {
		c.closeFuncs = append(c.closeFuncs, funcs...)
	}
}

// WithSignals add os signals to defaults SIGINT, SIGTERM
func WithSignals(sigs ...os.Signal) Option {
	return func(c *Closer) {
		c.sigs = append(c.sigs, sigs...)
	}
}
