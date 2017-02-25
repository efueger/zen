package zen

import (
	"log"
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
	namedSubRoutes map[string]*route
	regexSubRoutes map[string]*route
	params         map[int]string
}

// Route set handler for given pattern and method
func (s *Server) Route(method string, pattern string, handler HandlerFunc) {

	// add a named route
	if strings.Index(pattern, ":") == -1 {
		route := &route{
			namedSubRoutes: map[string]*route{},
			regexSubRoutes: map[string]*route{},
			handler:        handler,
		}
		k := strings.Join([]string{method, pattern}, "||")
		s.route.namedSubRoutes[k] = route
		return
	}

	// create tree route
	//split the url into sections
	parts := strings.Split(strings.TrimSuffix(strings.TrimPrefix(pattern, "/"), "/"), "/")

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

	s.route.generateRoute(method, parts, params, 0, handler)
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
		path := filepath.Clean(c.req.URL.Path)
		path = filepath.Join(dir, path)
		http.ServeFile(c.rw, c.req, path)
	})
}

// PProf serve golang's pprof tool
func (s *Server) PProf(pattern string) {
	s.Get(pattern, wrapHandler(pprof.Index))
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
			log.Printf("%+v", err)
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

	http.NotFound(c.rw, c.req)
}

// Filter adds the middleware filter.
func (s *Server) Filter(filter HandlerFunc) {
	s.filters = append(s.filters, filter)
}

// Required by http.Handler interface. This method is invoked by the
// http server and will handle all page routing
func (s *Server) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	log.Println("zen serveHTTP", r.RequestURI, r.URL.Path)
	w := &responseWriter{writer: rw}
	c := s.getContext(w, r)

	defer s.handlePanic(c)

	route := s.routeMatch(r.Method, r.URL.Path)
	if route != nil && route.handler != nil {
		route.parseParams(c)
		for _, f := range s.filters {
			f(c)
			if w.started {
				return
			}
		}
		route.handler(c)

		return
	}
	s.handleNotFound(c)
}

func (s *Server) routeMatch(method, pattern string) *route {
	k := generateKey(method, pattern)
	if r, ok := s.route.namedSubRoutes[k]; ok {
		return r
	}
	pattern = strings.TrimSuffix(strings.TrimPrefix(pattern, "/"), "/")

	parts := strings.Split(pattern, "/")
	return s.route.subRouteMatch(method, parts, 0)
}

func (r *route) subRouteMatch(method string, parts []string, index int) *route {
	if index >= len(parts) {
		return r
	}

	pattern := parts[index]
	k := generateKey(method, pattern)

	if sub, ok := r.namedSubRoutes[k]; ok {
		return sub.subRouteMatch(method, parts, index+1)
	}

	for k, v := range r.regexSubRoutes {
		if strings.HasPrefix(k, method) && v.regex.MatchString(pattern) {
			return v.subRouteMatch(method, parts, index+1)
		}
	}
	return nil

}

func (r *route) generateRoute(method string, parts []string, params map[int]string, index int, handler HandlerFunc) {
	if index >= len(parts) {
		r.params = params
		r.handler = handler
		return
	}

	pattern := parts[index]
	k := generateKey(method, pattern)

	var sub *route
	if _, ok := params[index]; ok {
		reg := regexp.MustCompile(pattern)
		sub = r.regexSubRoutes[k]
		if sub == nil {
			sub = &route{
				namedSubRoutes: map[string]*route{},
				regexSubRoutes: map[string]*route{},
				regex:          reg,
			}
			sub.generateRoute(method, parts, params, index+1, handler)
			r.regexSubRoutes[k] = sub
		}
	} else {
		sub = r.namedSubRoutes[k]
		if sub == nil {
			sub = &route{
				namedSubRoutes: map[string]*route{},
				regexSubRoutes: map[string]*route{},
			}
			sub.generateRoute(method, parts, params, index+1, handler)
			r.namedSubRoutes[k] = sub
		}
	}

}

func generateKey(method, pattern string) string {
	return strings.Join([]string{method, pattern}, "||")
}

func (r *route) parseParams(c *Context) {
	pattern := c.req.URL.Path
	pattern = strings.TrimSuffix(strings.TrimPrefix(pattern, "/"), "/")
	parts := strings.Split(pattern, "/")
	log.Println("parts", parts)
	log.Println(r.params)
	for i, k := range r.params {
		c.params[k] = parts[i]
	}

}
