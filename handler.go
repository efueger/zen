package zen

import (
	"net/http"
)

type (
	// HandlerFunc is a type alias for handler
	HandlerFunc func(*Context)
	// PanicHandler handle panic
	PanicHandler func(*Context, interface{})
)

func wrapF(h http.HandlerFunc) HandlerFunc {
	return func(c *Context) {
		h(c.rw, c.Req)
	}
}

func wrapH(h http.HandlerFunc) HandlerFunc {
	return func(c *Context) {
		h.ServeHTTP(c.rw, c.Req)
	}
}
