package init

import (
	"github.com/KennyMacCormik/common/gin_factory"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"github.com/KennyMacCormik/otel/backend/internal/http/handlers/storage"
	"github.com/KennyMacCormik/otel/backend/pkg/cache"

	httpWithGin "github.com/KennyMacCormik/otel/backend/pkg/gin"
	"github.com/KennyMacCormik/otel/backend/pkg/gin/gin_rate_limiter"
	"github.com/KennyMacCormik/otel/backend/pkg/gin/gin_request_id"

	"github.com/KennyMacCormik/otel/backend/internal/http/middleware"
)

const otelGinMiddlewareName = "api"

func HttpServer(conf *Config, st cache.CacheInterface) *httpWithGin.Server {
	return httpWithGin.NewHttpServer(
		conf.Http.Endpoint,
		initRouter(conf, st),
		conf.Http.ReadTimeout,
		conf.Http.ReadTimeout,
		conf.Http.ReadTimeout,
	)
}

func initRouter(conf *Config, st cache.CacheInterface) *gin_factory.GinFactory {
	ginFactory := gin_factory.NewGinFactory()

	ginFactory.AddMiddleware(
		middleware.GetTraceParent(),
		otelgin.Middleware(otelGinMiddlewareName),
		gin_request_id.RequestIDMiddleware(),
		gin_rate_limiter.NewRateLimiter(conf.RateLimiter.MaxRunning,
			conf.RateLimiter.MaxWait, conf.RateLimiter.RetryAfter).GetRateLimiter(),
	)

	ginFactory.AddHandlers(storage.NewStorageHandlers(st))

	return ginFactory
}
