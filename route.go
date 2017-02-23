package zen

import (
	"net/http"
	"net/http/pprof"
	"net/url"
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
	method  string
	regex   *regexp.Regexp
	params  map[int]string
	handler HandlerFunc
}

// Route set handler for given pattern and method
func (s *Server) Route(method string, pattern string, handler HandlerFunc) {
	s.AddRoute(method, pattern, handler)
}

// Get adds a new Route for GET requests.
func (s *Server) Get(pattern string, handler HandlerFunc) {
	s.AddRoute(GET, pattern, handler)
}

// Put adds a new Route for PUT requests.
func (s *Server) Put(pattern string, handler HandlerFunc) {
	s.AddRoute(PUT, pattern, handler)
}

// Del adds a new Route for DELETE requests.
func (s *Server) Del(pattern string, handler HandlerFunc) {
	s.AddRoute(DELETE, pattern, handler)
}

// Patch adds a new Route for PATCH requests.
func (s *Server) Patch(pattern string, handler HandlerFunc) {
	s.AddRoute(PATCH, pattern, handler)
}

// Post adds a new Route for POST requests.
func (s *Server) Post(pattern string, handler HandlerFunc) {
	s.AddRoute(POST, pattern, handler)
}

// Static :Adds a new Route for Static http requests. Serves
// static files from the specified directory
func (s *Server) Static(pattern string, dir string) {
	//append a regex to the param to match everything
	// that comes after the prefix
	pattern = pattern + "(.+)"
	s.AddRoute(GET, pattern, func(c *Context) {
		path := filepath.Clean(c.req.URL.Path)
		path = filepath.Join(dir, path)
		http.ServeFile(c.rw, c.req, path)
	})
}

// Pprof serve golang's pprof tool
func (s *Server) Pprof(pattern string) {
	s.Get(pattern, wrapHandler(pprof.Index))
}

// AddRoute : Adds a new Route to the Handler
func (s *Server) AddRoute(method string, pattern string, handler HandlerFunc) {

	//split the url into sections
	parts := strings.Split(pattern, "/")

	//find params that start with ":"
	//replace with regular expressions
	j := 0
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
			params[j] = part
			parts[i] = expr
			j++
		}
	}

	//recreate the url pattern, with parameters replaced
	//by regular expressions. then compile the regex
	pattern = strings.Join(parts, "/")
	regex, regexErr := regexp.Compile(pattern)
	if regexErr != nil {
		// fail earlier
		panic(regexErr)
	}

	//now create the Route
	route := &route{}
	route.method = method
	route.regex = regex
	route.handler = handler
	route.params = params

	//and finally append to the list of Routes
	s.routes = append(s.routes, route)
}

// Filter adds the middleware filter.
func (s *Server) Filter(filter HandlerFunc) {
	s.filters = append(s.filters, filter)
}

// FilterParam adds the middleware filter if the REST URL parameter exists.
func (s *Server) FilterParam(param string, filter HandlerFunc) {
	if !strings.HasPrefix(param, ":") {
		param = ":" + param
	}

	s.Filter(func(c *Context) {
		p := c.req.URL.Query().Get(param)
		if len(p) > 0 {
			filter(c)
		}
	})
}

// Required by http.Handler interface. This method is invoked by the
// http server and will handle all page routing
func (s *Server) ServeHTTP(rw http.ResponseWriter, r *http.Request) {

	requestPath := r.URL.Path

	c := s.contextPool.Get().(*Context)
	c.req = r
	c.rw = rw

	defer func() {
		if e := recover(); e != nil {
			if s.PanicHandler != nil {
				s.PanicHandler(c, e)
			} else {
				http.Error(rw, "internal server error", http.StatusInternalServerError)
			}
		}
	}()

	//find a matching Route
	for _, route := range s.routes {

		//if the methods don't match, skip this handler
		//i.e if request.Method is 'PUT' Route.Method must be 'PUT'
		if r.Method != route.method {
			continue
		}

		//check if Route pattern matches url
		if !route.regex.MatchString(requestPath) {
			continue
		}

		//get submatches (params)
		matches := route.regex.FindStringSubmatch(requestPath)

		//double check that the Route matches the URL pattern.
		if len(matches[0]) != len(requestPath) {
			continue
		}

		if len(route.params) > 0 {
			//add url parameters to the query param map
			values := r.URL.Query()
			for i, match := range matches[1:] {
				values.Add(route.params[i], match)
			}

			//reassemble query params and add to RawQuery
			r.URL.RawQuery = url.Values(values).Encode() + "&" + r.URL.RawQuery
			//r.URL.RawQuery = url.Values(values).Encode()
		}

		//execute middleware filters
		for _, filter := range s.filters {
			filter(c)
		}

		//Invoke the request handler
		route.handler(c)
		return
	}

	if s.NotFoundHandler != nil {
		s.NotFoundHandler(c)
	} else {
		http.NotFound(rw, r)
	}
}
