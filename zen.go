package zen

import (
	"net/http"
	"sync"
	"time"
)

const (
	// Version is current version num
	Version = "v1.0.0Beta"
)

type (
	// Server struct
	Server struct {
		rTimeOut    time.Duration
		wTimeOut    time.Duration
		contextPool *sync.Pool
	}
)

// NewServer will create a Server instance and response with a pointer which point to it
func NewServer() *Server {
	s := &Server{}
	s.contextPool.New = func() interface{} {
		c := Context{}
		return &c
	}
	return s
}

// Run server
func (s *Server) Run(addr string) error {
	return http.ListenAndServe(addr, s)
}

// RunTLS Run server with tls
func (s *Server) RunTLS(addr string, certFile string, keyFile string) error {
	return http.ListenAndServeTLS(addr, certFile, keyFile, s)
}

func (s *Server) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	c := s.contextPool.Get().(*Context)
	s.serveHTTPRequest(c)
}

func (s *Server) serveHTTPRequest(c *Context) {

}
