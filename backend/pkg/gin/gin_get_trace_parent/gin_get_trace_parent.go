package gin_get_trace_parent

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

func GetTraceParent() gin.HandlerFunc {
	return func(c *gin.Context) {
		propagator := otel.GetTextMapPropagator()
		ctx := propagator.Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))

		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
