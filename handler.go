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

// wrapF wrap a http handlerfunc into HandlerFunc
func wrapF(h http.HandlerFunc) HandlerFunc {
	return func(c *Context) {
		h(c.rw, c.Req)
	}
}
