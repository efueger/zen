package zen

import (
	"net/http"
)

type (
	// HandlerFunc is a type alias for handler
	HandlerFunc func(*Context)
)

func wrapHandler(h http.HandlerFunc) HandlerFunc {
	return func(c *Context) {
		h(c.rw, c.req)
	}
}
