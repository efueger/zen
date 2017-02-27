package zen

import (
	"sync"
	"testing"
)

func TestServer_routeMatch(t *testing.T) {
	type fields struct {
		routeTree       map[string]*route
		notFoundHandler HandlerFunc
		panicHandler    PanicHandler
		filters         []HandlerFunc
		contextPool     *sync.Pool
	}
	type args struct {
		method  string
		pattern string
	}

	tests := []struct {
		name     string
		fields   fields
		routes   []args
		args     args
		wantNil1 bool
		wantNil2 bool
	}{
		{"case1",
			fields{
				routeTree: map[string]*route{},
				contextPool: &sync.Pool{
					New: func() interface{} {
						c := Context{
							params: map[string]string{},
							rw:     &responseWriter{},
						}
						return &c
					},
				},
			},
			[]args{
				{
					"GET", "/users/test",
				},
				{
					"POST", "/users/test",
				},
			},
			args{
				"GET", "/users/test",
			},
			false,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				routeTree:       tt.fields.routeTree,
				notFoundHandler: tt.fields.notFoundHandler,
				panicHandler:    tt.fields.panicHandler,
				filters:         tt.fields.filters,
				contextPool:     *tt.fields.contextPool,
			}
			for _, r := range tt.routes {
				s.Route(r.method, r.pattern, nil)
			}
			got1, got2 := s.routeMatch(tt.args.method, tt.args.pattern)
			if (got1 == nil) != tt.wantNil1 {
				t.Errorf("Server.routeMatch() got1 = %v, want nil? %v", got1, tt.wantNil1)
			}
			if (got2 == nil) != tt.wantNil2 {
				t.Errorf("Server.routeMatch() got2 = %v, want nil? %v", got2, tt.wantNil2)
			}
		})
	}
}
