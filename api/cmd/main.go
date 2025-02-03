package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/KennyMacCormik/common/log"
	"github.com/KennyMacCormik/otel/backend/pkg/otel"

	initApp "github.com/KennyMacCormik/otel/api/internal/init"

	"github.com/KennyMacCormik/otel/api/internal/cache"
	"github.com/KennyMacCormik/otel/api/internal/client"
	"github.com/KennyMacCormik/otel/api/internal/compute"
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

	// init cache
	c, err := cache.NewCache(lg)
	if err != nil {
		lg.Error("failed to initialize cache", "error", err)
		os.Exit(errExit)
	}
	defer func() {
		err = c.Close(context.Background())
		if err != nil {
			lg.Error("failed to close cache", "error", err)
		} else {
			lg.Info("cache closed")
		}
	}()
	// init http client
	httpClient := client.NewClient(conf.Client.BackendEndpoint, conf.Client.BackendRequestTimeout)
	// init compute
	comp := compute.NewComputeLayer(c, httpClient, lg)
	// init server
	closer := myinit.InitServer(conf, errExit, comp, lg)
	defer closer()
	// gracefully shutting down
	quit := make(chan os.Signal, 1)
	defer close(quit)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)
	<-quit
}

func gracefulStop() {
	_ = syscall.Kill(os.Getpid(), syscall.SIGTERM)
}
