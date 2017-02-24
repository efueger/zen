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
		http.Server
		contextPool     sync.Pool
		route           *route
		notFoundHandler HandlerFunc
		panicHandler    PanicHandler
		filters         []HandlerFunc
	}
)

// NewServer will create a Server instance and response with a pointer which point to it
func NewServer() *Server {
	// create root router
	route := &route{namedSubRoutes: map[string]*route{}, regexSubRoutes: map[string]*route{}}

	s := &Server{route: route, contextPool: sync.Pool{}, filters: []HandlerFunc{}}
	s.contextPool.New = func() interface{} {
		c := Context{}
		return &c
	}
	return s
}

// Run server
func (s *Server) Run(addr string) error {
	log.Println("start zen on", addr)
	s.Addr = addr
	s.Handler = s
	return s.ListenAndServe()
}

// RunTLS Run server with tls
func (s *Server) RunTLS(addr string, certFile string, keyFile string) error {
	s.Addr = addr
	s.Handler = s
	return s.ListenAndServeTLS(certFile, keyFile)
}
