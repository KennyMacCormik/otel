package helpers

import (
	"context"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func StartSpanWithCtx(ctx context.Context, traceName, spanName string) (context.Context, trace.Span) {
	tracer := otel.Tracer(traceName)
	newCtx, span := tracer.Start(ctx, spanName)
	return newCtx, span
}

func StartSpanWithGinCtx(c *gin.Context, traceName, spanName string) trace.Span {
	tracer := otel.Tracer(traceName)
	ctx, span := tracer.Start(c.Request.Context(), spanName)
	c.Request = c.Request.WithContext(ctx)
	return span
}

func SetSpanErr(span trace.Span, err error) {
	span.SetStatus(codes.Error, err.Error())
	span.RecordError(err)
	span.SetAttributes(attribute.String("error.message", err.Error()))
}
