package zen

import (
	"net/http"
	"net/http/pprof"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	// CONNECT : http method CONNECT
	CONNECT = "CONNECT"
	// DELETE : http method DELETE
	DELETE = "DELETE"
	// GET : http method GET
	GET = "GET"
	// HEAD : http method HEAD
	HEAD = "HEAD"
	// OPTIONS : http method OPTIONS
	OPTIONS = "OPTIONS"
	// PATCH : http method PATCH
	PATCH = "PATCH"
	// POST : http method POST
	POST = "POST"
	// PUT : http method PUT
	PUT = "PUT"
	// TRACE : http method TRACE
	TRACE = "TRACE"
)

type route struct {
	regex          *regexp.Regexp
	handler        HandlerFunc
	children       []*route
	namedSubRoutes map[string]*route
	params         map[int]string
}

func (s *Server) methodRouteTree(method string) *route {
	r, ok := s.routeTree[method]
	if !ok {
		s.routeTree[method] = &route{
			namedSubRoutes: map[string]*route{},
		}
		r = s.routeTree[method]
	}
	return r
}

// Route set handler for given pattern and method
func (s *Server) Route(method string, pattern string, handler HandlerFunc) {
	// assert method and pattern is not empty, else panic
	assert(len(method) > 0, "Method should not be empty")
	assert(len(pattern) > 0 && pattern[0] == '/', "Pattern should start with /")

	// get method route tree
	r := s.methodRouteTree(method)

	// add a named route
	if strings.Index(pattern, ":") == -1 {
		r.namedSubRoutes[pattern] = &route{
			handler: handler,
		}
		return
	}

	// create tree route
	//split the url into sections
	parts := strings.Split(pattern, "/")

	//find params that start with ":"
	//replace with regular expressions
	params := make(map[int]string)
	for i, part := range parts {
		if strings.HasPrefix(part, ":") {
			expr := "([^/]+)"
			//a user may choose to override the defult expression
			// similar to expressjs: ‘/user/:id([0-9]+)’
			if index := strings.Index(part, "("); index != -1 {
				expr = part[index:]
				part = part[:index]
			}
			params[i] = part
			parts[i] = expr
		}
	}

	r.generateRoute(parts, params, 0, handler)
}

// Get adds a new Route for GET requests.
func (s *Server) Get(pattern string, handler HandlerFunc) {
	s.Route(GET, pattern, handler)
}

// Put adds a new Route for PUT requests.
func (s *Server) Put(pattern string, handler HandlerFunc) {
	s.Route(PUT, pattern, handler)
}

// Del adds a new Route for DELETE requests.
func (s *Server) Del(pattern string, handler HandlerFunc) {
	s.Route(DELETE, pattern, handler)
}

// Patch adds a new Route for PATCH requests.
func (s *Server) Patch(pattern string, handler HandlerFunc) {
	s.Route(PATCH, pattern, handler)
}

// Post adds a new Route for POST requests.
func (s *Server) Post(pattern string, handler HandlerFunc) {
	s.Route(POST, pattern, handler)
}

// Static :Adds a new Route for Static http requests. Serves
// static files from the specified directory
func (s *Server) Static(pattern string, dir string) {
	//append a regex to the param to match everything
	// that comes after the prefix
	pattern = pattern + "(.+)"
	s.Route(GET, pattern, func(c *Context) {
		path := filepath.Clean(c.Req.URL.Path)
		path = filepath.Join(dir, path)
		http.ServeFile(c.rw, c.Req, path)
	})
}

// PProf serve golang's pprof tool
func (s *Server) PProf(pattern string) {
	s.Get(pattern, wrapF(pprof.Index))
}

// HandleNotFound set server's notFoundHandler
func (s *Server) HandleNotFound(handler HandlerFunc) {
	s.notFoundHandler = handler
}

// HandlePanic set server's panicHandler
func (s *Server) HandlePanic(handler PanicHandler) {
	s.panicHandler = handler
}

// handlePanic call server's panic handler
func (s *Server) handlePanic(c *Context) {

	if err := recover(); err != nil {
		if s.panicHandler != nil {
			s.panicHandler(c, err)
		} else {
			http.Error(c.rw, "internal server error", http.StatusInternalServerError)
		}
	}
}

// handleNotFound call server's not found handler
func (s *Server) handleNotFound(c *Context) {

	if s.notFoundHandler != nil {
		s.notFoundHandler(c)
		return
	}

	http.NotFound(c.rw, c.Req)
}

func (s *Server) routeMatch(method, pattern string) (*route, []string) {
	parts := strings.Split(pattern, "/")
	r := s.routeTree[method]
	if r == nil {
		return nil, parts
	}

	if r, ok := r.namedSubRoutes[pattern]; ok {
		return r, parts
	}

	return r.subRouteMatch(parts, 0), parts
}

func (r *route) subRouteMatch(parts []string, index int) *route {
	if index >= len(parts) {
		return r
	}

	pattern := parts[index]

	if sub, ok := r.namedSubRoutes[pattern]; ok {
		return sub.subRouteMatch(parts, index+1)
	}

	for _, v := range r.children {
		if v.regex.MatchString(pattern) {
			return v.subRouteMatch(parts, index+1)
		}
	}
	return nil

}

func (r *route) getSubRoute(pattern string) *route {
	for _, route := range r.children {
		if route.regex.MatchString(pattern) {
			return route
		}
	}
	return nil
}

func (r *route) generateRoute(parts []string, params map[int]string, index int, handler HandlerFunc) {
	if index >= len(parts) {
		r.params = params
		r.handler = handler
		return
	}

	pattern := parts[index]

	var sub *route
	if _, ok := params[index]; ok {
		reg := regexp.MustCompile(pattern)
		sub = r.getSubRoute(pattern)
		if sub == nil {
			sub = &route{
				namedSubRoutes: map[string]*route{},
				regex:          reg,
			}
			r.children = append(r.children, sub)
		}
		sub.generateRoute(parts, params, index+1, handler)

	} else {
		sub = r.namedSubRoutes[pattern]
		if sub == nil {
			sub = &route{
				namedSubRoutes: map[string]*route{},
			}
			r.namedSubRoutes[pattern] = sub
		}
		sub.generateRoute(parts, params, index+1, handler)
	}

}
