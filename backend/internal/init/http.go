package init

import (
	"github.com/KennyMacCormik/HerdMaster/pkg/cache"
	"github.com/KennyMacCormik/HerdMaster/pkg/gin/router"
	"github.com/KennyMacCormik/otel/backend/internal/http/handlers"
	"github.com/KennyMacCormik/otel/backend/internal/http/middleware"
	httpWithGin "github.com/KennyMacCormik/otel/otel-common/gin"
	"github.com/KennyMacCormik/otel/otel-common/gin/gin_rate_limiter"
	"github.com/KennyMacCormik/otel/otel-common/gin/gin_request_id"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

const otelGinMiddlewareName = "api"

func HttpServer(conf *Config, st cache.Interface) *httpWithGin.Server {
	return httpWithGin.NewHttpServer(
		conf.Http.Endpoint,
		initRouter(conf, st),
		conf.Http.ReadTimeout,
		conf.Http.ReadTimeout,
		conf.Http.ReadTimeout,
	)
}

func initRouter(conf *Config, st cache.Interface) *router.GinFactory {
	ginFactory := router.NewGinFactory()
	ginFactory.AddMiddleware(
		middleware.GetTraceParent(),
		otelgin.Middleware(otelGinMiddlewareName),
		gin_request_id.RequestIDMiddleware(),
		gin_rate_limiter.NewRateLimiter(conf.RateLimiter.MaxRunning,
			conf.RateLimiter.MaxWait, conf.RateLimiter.RetryAfter).GetRateLimiter(),
	)
	ginFactory.AddHandlers(handlers.NewStorageHandlers(st))
	return ginFactory
}
