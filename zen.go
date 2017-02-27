package zen

import (
	"log"
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
		routeTree       map[string]*route
		notFoundHandler HandlerFunc
		panicHandler    PanicHandler
		filters         []HandlerFunc
		contextPool     sync.Pool
	}
)

// NewServer will create a Server instance and response with a pointer which point to it
func NewServer() *Server {
	// create root router

	s := &Server{routeTree: map[string]*route{}, contextPool: sync.Pool{}, filters: []HandlerFunc{}}
	s.contextPool.New = func() interface{} {
		c := Context{params: map[string]string{}, rw: &responseWriter{}}
		return &c
	}
	return s
}

// Required by http.Handler interface. This method is invoked by the
// http server and will handle all page routing
func (s *Server) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	// warp response writer

	c := s.getContext(rw, r)
	// c.parseInput()
	// put context into pool
	defer s.putBackContext(c)
	// handle panic
	// defer s.handlePanic(c)

	route, parts := s.routeMatch(r.Method, r.RequestURI)
	if route != nil && route.handler != nil {
		route.parseParams(c, parts)
		for _, f := range s.filters {
			f(c)
			if c.rw.written {
				return
			}
		}
		route.handler(c)
		return
	}

	s.handleNotFound(c)
}

// Run server on addr
func (s *Server) Run(addr string) error {
	log.Println("start zen on", addr)

	return http.ListenAndServe(addr, s)
}

// RunTLS Run server on addr with tls
func (s *Server) RunTLS(addr string, certFile string, keyFile string) error {
	log.Println("start zen with tls on", addr)

	return http.ListenAndServeTLS(addr, certFile, keyFile, s)
}
