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

func wrapHandler(h http.HandlerFunc) HandlerFunc {
	return func(c *Context) {
		h(c.rw, c.req)
	}
}

func (h HandlerFunc) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	c := Context{}
	c.req = req
	c.rw = rw
	h(&c)
}
