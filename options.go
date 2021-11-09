package closer

import (
	"io"
	"os"
)

//WithClosers .
func WithClosers(closers ...io.Closer) Option {
	return func(c *Closer) {
		c.closers = append(c.closers, closers...)
	}
}

//WithCloseFuncs .
func WithCloseFuncs(funcs ...func()) Option {
	return func(c *Closer) {
		c.closeFuncs = append(c.closeFuncs, funcs...)
	}
}

//WithSignals add signals to defaults SIGINT, SIGTERM
func WithSignals(sigs ...os.Signal) Option {
	return func(c *Closer) {
		c.sigs = append(c.sigs, sigs...)
	}
}
