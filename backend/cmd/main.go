package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/KennyMacCormik/common/log"
	initApp "github.com/KennyMacCormik/otel/backend/internal/init"
	"github.com/KennyMacCormik/otel/backend/internal/storage"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

const errExit = 1

func main() {
	conf := initApp.GetConfig()
	if conf == nil {
		log.Error("failed to initialize config")
		os.Exit(errExit)

	}

	log.Configure(log.WithLogLevel(conf.Log.Level))

	log.Info("config initialized")
	log.Debug(fmt.Sprintf("%+v", conf))

	tp, err := initApp.OTelInit(context.Background(), conf)
	if err != nil {
		log.Error("failed to initialize OTel", "error", err)
		os.Exit(errExit)
	}
	log.Info("OTel initialized")
	defer func() {
		ctxStop, cancel := context.WithTimeout(context.Background(), conf.OTel.ShutdownTimeout)
		defer cancel()
		err = tp.Shutdown(ctxStop)
		if err != nil {
			log.Error("failed to shutdown OTel", "error", err)
		}
	}()

	st := storage.NewStorage()
	log.Info("cache initialized")

	httpSvr := initApp.HttpServer(conf, st)
	log.Info("http server initialized")
	defer func() {
		err = httpSvr.Close(conf.Http.ShutdownTimeout)
		if err != nil {
			log.Error("failed to shutdown http server", "error", err)
		}
	}()

	go func() {
		err = httpSvr.Start()
		if err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				log.Error("Failed to start server", "error", err)
				os.Exit(errExit)
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
