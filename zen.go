package zen

import (
	"net/http"
	"sync"
)

const (
	// Version is current version num
	Version = "v1.0.0Beta"
)

type (
	// Server struct
	Server struct {
		routeTree       []*methodNode
		notFoundHandler HandlerFunc
		panicHandler    PanicHandler
		filters         []HandlerFunc
		contextPool     sync.Pool
	}
)

// New will create a Server instance and return a pointer which point to it
func New() *Server {

	s := &Server{contextPool: sync.Pool{}, filters: []HandlerFunc{}}
	s.contextPool.New = func() interface{} {
		c := Context{rw: &responseWriter{}}
		return &c
	}
	return s
}

// Required by http.Handler interface. This method is invoked by the
// http server and will handle all page routing
func (s *Server) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	// get context instance from pool
	c := s.getContext(rw, r)
	// put context back into pool
	defer s.putBackContext(c)
	// handle panic
	defer s.handlePanic(c)

	httpMethod := r.Method
	path := r.URL.Path

	for i := 0; i < len(s.routeTree); i++ {
		t := s.routeTree[i]
		if t.method == httpMethod {
			handlers, params := t.node.get(path, c.params)
			if handlers != nil {
				c.params = params
				for _, h := range handlers {
					h(c)
					if c.rw.written {
						return
					}
				}
			}
		}
	}

	s.handleNotFound(c)
}

// Run server on addr
func (s *Server) Run(addr string) error {
	return http.ListenAndServe(addr, s)
}

// RunTLS Run server on addr with tls
func (s *Server) RunTLS(addr string, certFile string, keyFile string) error {
	return http.ListenAndServeTLS(addr, certFile, keyFile, s)
}
