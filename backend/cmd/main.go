package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/KennyMacCormik/common/log"

	initApp "github.com/KennyMacCormik/otel/backend/internal/init"
	"github.com/KennyMacCormik/otel/backend/internal/storage"
	"github.com/KennyMacCormik/otel/backend/pkg/otel"
)

const (
	otelServiceName = "backend"
	errExit         = 1
)

func main() {
	conf := initApp.GetConfig()
	if conf == nil {
		log.Error("failed to initialize config")
		os.Exit(errExit)
	}

	if conf.Log.Format == "json" {
		log.Configure(log.WithLogLevel(conf.Log.Level), log.WithJSONFormat())
	} else {
		log.Configure(log.WithLogLevel(conf.Log.Level), log.WithTextFormat())
	}

	log.Info("config initialized")
	log.Debug(fmt.Sprintf("%+v", conf))

	tp, err := otel.OTelInit(context.Background(), conf.OTel.Endpoint, otelServiceName)
	if err != nil {
		log.Error("failed to initialize OTel", "error", err)
		gracefulStop()
	}
	log.Info("OTel initialized")
	defer func() {
		ctxStop, cancel := context.WithTimeout(context.Background(), conf.OTel.ShutdownTimeout)
		defer cancel()
		err = tp.Shutdown(ctxStop)
		if err != nil {
			log.Warn("failed to shutdown OTel", "error", err)
		}
	}()

	st := storage.NewStorage()
	log.Info("cache initialized")

	httpSvr := initApp.HttpServer(conf, st)
	log.Info("http server initialized")
	defer func() {
		err = httpSvr.Close(conf.Http.ShutdownTimeout)
		if err != nil {
			log.Warn("failed to shutdown http server", "error", err)
		}
	}()

	go func() {
		err = httpSvr.Start()
		if err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				log.Error("Failed to start server", "error", err)
				gracefulStop()
			}
		}
	}()
	log.Info("server started")

	quit := make(chan os.Signal, 1)
	defer close(quit)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)
	<-quit
	log.Info("server stopped")
}

func gracefulStop() {
	_ = syscall.Kill(os.Getpid(), syscall.SIGTERM)
}
