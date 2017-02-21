package zen

import (
	"path"
	"strings"
)

type (
	route struct {
		path    string
		method  string
		handler string
	}
)

// Route set handler for given route and method
func (s *Server) Route(method, route string, handler HandlerFunc) {
	key := path.Clean(route) + "::" + strings.ToUpper(method)
	s.routes[key] = handler
}
