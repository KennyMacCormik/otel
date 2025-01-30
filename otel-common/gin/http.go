package http

import (
	"context"
	"github.com/KennyMacCormik/HerdMaster/pkg/gin/router"
	"net/http"
	"time"
)

type HttpServer struct {
	svr http.Server
}

func NewHttpServer(hostPort string, r *router.GinFactory, rTimeout time.Duration,
	wTimeout time.Duration, iTimeout time.Duration) *HttpServer {
	return &HttpServer{
		svr: http.Server{
			Addr:         hostPort,
			Handler:      r.CreateRouter(),
			ReadTimeout:  rTimeout,
			WriteTimeout: wTimeout,
			IdleTimeout:  iTimeout,
		},
	}
}

func (s *HttpServer) Start() error {
	return s.svr.ListenAndServe()
}

func (s *HttpServer) Close(t time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), t)
	defer cancel()
	return s.svr.Shutdown(ctx)
}
