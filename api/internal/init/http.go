package init

import (
	"api/internal/compute"
	"api/internal/http/handlers"
	"errors"
	myhttp "github.com/KennyMacCormik/HerdMaster/pkg/gin"
	"github.com/KennyMacCormik/HerdMaster/pkg/gin/middleware"
	"github.com/KennyMacCormik/HerdMaster/pkg/gin/router"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"log/slog"
	"net/http"
	"os"
	"strconv"
)

const otelGinMiddlewareName = "api"

func InitServer(conf *Config, errExit int, comp compute.Interface, lg *slog.Logger) (closer func()) {
	svr := myhttp.NewHttpServer(
		conf.Http.Host+":"+strconv.Itoa(conf.Http.Port),
		initRouter(conf, comp, lg),
		conf.Http.ReadTimeout,
		conf.Http.ReadTimeout,
		conf.Http.ReadTimeout,
	)
	go func() {
		lg.Info("Starting http server")
		err := svr.Start()
		if err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				lg.Error("Failed to start server", "error", err)
				os.Exit(errExit)
			}
		}
	}()
	return func() {
		err := svr.Close(conf.Http.ShutdownTimeout)
		if err != nil {
			lg.Error("Failed to gracefully stop server", "error", err)
			return
		}
		lg.Info("Server gracefully stopped")
	}
}

func initRouter(conf *Config, comp compute.Interface, lg *slog.Logger) *router.GinFactory {
	ginFactory := router.NewGinFactory()
	ginFactory.AddMiddleware(
		otelgin.Middleware(otelGinMiddlewareName),
		middleware.RequestIDMiddleware(),
		middleware.NewRateLimiter(conf.RateLimiter.MaxRunning,
			conf.RateLimiter.MaxWait, conf.RateLimiter.RetryAfter, lg).GetRateLimiter(),
	)
	ginFactory.AddHandlers(handlers.NewStorageHandlers(comp, lg))
	return ginFactory
}
