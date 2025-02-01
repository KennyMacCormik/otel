package http

import (
	"context"
	"net/http"
	"time"

	"github.com/KennyMacCormik/common/gin_factory"
)

type Server struct {
	svr http.Server
}

func NewHttpServer(endpoint string, r *gin_factory.GinFactory, rTimeout time.Duration,
	wTimeout time.Duration, iTimeout time.Duration) *Server {
	return &Server{
		svr: http.Server{
			Addr:         endpoint,
			Handler:      r.CreateRouter(),
			ReadTimeout:  rTimeout,
			WriteTimeout: wTimeout,
			IdleTimeout:  iTimeout,
		},
	}
}

func (s *Server) Start() error {
	return s.svr.ListenAndServe()
}

func (s *Server) Close(t time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), t)
	defer cancel()
	return s.svr.Shutdown(ctx)
}
