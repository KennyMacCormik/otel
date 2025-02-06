package init

import (
	"github.com/KennyMacCormik/common/gin_factory"
	httpWithGin "github.com/KennyMacCormik/otel/backend/pkg/gin"
	"github.com/KennyMacCormik/otel/backend/pkg/gin/gin_rate_limiter"
	"github.com/KennyMacCormik/otel/backend/pkg/gin/gin_request_id"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	storageHandlers "github.com/KennyMacCormik/otel/api/internal/http/handlers/storage"
	"github.com/KennyMacCormik/otel/api/internal/service"
)

const otelGinMiddlewareName = "api"

func InitServer(conf *Config, svc service.ServiceInterface) *httpWithGin.GinServer {
	return httpWithGin.NewHttpServer(
		conf.Http.Endpoint,
		initRouter(conf, svc),
		conf.Http.ReadTimeout,
		conf.Http.ReadTimeout,
		conf.Http.ReadTimeout,
	)
}

func initRouter(conf *Config, svc service.ServiceInterface) *gin_factory.GinFactory {
	ginFactory := gin_factory.NewGinFactory()

	ginFactory.AddMiddleware(
		otelgin.Middleware(otelGinMiddlewareName),
		gin_request_id.RequestIDMiddleware(),
		gin_rate_limiter.NewRateLimiter(
			conf.RateLimiter.MaxRunning,
			conf.RateLimiter.MaxWait,
			conf.RateLimiter.RetryAfter,
		).GetRateLimiter(),
	)

	ginFactory.AddHandlers(storageHandlers.NewStorageHandler(svc).GetGinHandler())

	return ginFactory
}
