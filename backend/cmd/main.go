package main

import (
	myinit "backend/internal/init"
	"backend/internal/storage"
	"context"
	"github.com/KennyMacCormik/HerdMaster/pkg/log"
	"github.com/KennyMacCormik/HerdMaster/pkg/val"
	"os"
	"os/signal"
	"syscall"
)

const errExit = 1

func main() {
	// init default logger, errors ignored as per documentation
	lg, _ := log.GetLogger()
	// init validator
	v := val.GetValidator()
	// init conf
	conf, err := myinit.InitConfig(v)
	if err != nil {
		lg.Error("failed to initialize config", "error", err)
		os.Exit(errExit)
	}
	// reconfigure logger according to loaded conf
	lg, err = log.ConfigureLogger(log.WithConfig(conf.Log.Level, conf.Log.Format))
	if err != nil {
		lg.Warn("failed to configure logger", "error", err)
	}
	lg.Debug("initialized config", "config", conf)
	// init trace
	stopTrace, err := myinit.InitOtel(context.Background(), conf, lg)
	if err != nil {
		lg.Error("failed to initialize otel", "error", err)
		os.Exit(errExit)
	}
	defer func() { _ = stopTrace() }()
	// init storage
	st := storage.NewStorage()
	// init server
	closer := myinit.InitServer(conf, errExit, st, lg)
	defer closer()
	// gracefully shutting down
	quit := make(chan os.Signal, 1)
	defer close(quit)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)
	<-quit
}
