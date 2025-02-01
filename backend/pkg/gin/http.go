package http

import (
	"context"
	"net/http"
	"time"

	"github.com/KennyMacCormik/common/gin_factory"
)

type GinServer struct {
	svr http.Server
}

func NewHttpServer(endpoint string, r *gin_factory.GinFactory, rTimeout time.Duration,
	wTimeout time.Duration, iTimeout time.Duration) *GinServer {
	return &GinServer{
		svr: http.Server{
			Addr:         endpoint,
			Handler:      r.CreateRouter(),
			ReadTimeout:  rTimeout,
			WriteTimeout: wTimeout,
			IdleTimeout:  iTimeout,
		},
	}
}

func (s *GinServer) Start() error {
	return s.svr.ListenAndServe()
}

func (s *GinServer) Close(t time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), t)
	defer cancel()

	return s.svr.Shutdown(ctx)
}
