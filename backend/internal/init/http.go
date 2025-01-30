package init

import (
	"errors"
	"github.com/KennyMacCormik/HerdMaster/pkg/cache"
	myhttp "github.com/KennyMacCormik/HerdMaster/pkg/gin"
	"github.com/KennyMacCormik/HerdMaster/pkg/gin/router"
	"github.com/KennyMacCormik/otel/backend/internal/http/handlers"
	"github.com/KennyMacCormik/otel/backend/internal/http/middleware"
	"github.com/KennyMacCormik/otel/otel-common/gin/gin_rate_limiter"
	"github.com/KennyMacCormik/otel/otel-common/gin/gin_request_id"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"log/slog"
	"net/http"
	"os"
	"strconv"
)

const otelGinMiddlewareName = "api"

func InitServer(conf *Config, errExit int, st cache.Interface, lg *slog.Logger) (closer func()) {
	svr := myhttp.NewHttpServer(
		conf.Http.Host+":"+strconv.Itoa(conf.Http.Port),
		initRouter(conf, st, lg),
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

func initRouter(conf *Config, st cache.Interface, lg *slog.Logger) *router.GinFactory {
	ginFactory := router.NewGinFactory()
	ginFactory.AddMiddleware(
		middleware.GetTraceParent(),
		otelgin.Middleware(otelGinMiddlewareName),
		gin_request_id.RequestIDMiddleware(),
		gin_rate_limiter.NewRateLimiter(int64(conf.RateLimiter.MaxRunning),
			int64(conf.RateLimiter.MaxWait), int64(conf.RateLimiter.RetryAfter)).GetRateLimiter(),
	)
	ginFactory.AddHandlers(handlers.NewStorageHandlers(st, lg))
	return ginFactory
}
