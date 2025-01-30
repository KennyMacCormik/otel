package init

import (
	"context"
	"fmt"
	"github.com/KennyMacCormik/HerdMaster/pkg/conv"
	customLogger "github.com/KennyMacCormik/common/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
	"log"
	"log/slog"
)

const otelServiceName = "backend"

// CustomWriter redirects log output to slog. Created only to intercept otel error messages
type CustomWriter struct {
	lg *slog.Logger
}

func NewCustomWriter(lg *slog.Logger) *CustomWriter {
	return &CustomWriter{lg: lg}
}

func (cw *CustomWriter) RedirectLoggerToSlog() {
	log.SetFlags(0) // Remove timestamps from standard log
	log.SetOutput(cw)
}

func (cw *CustomWriter) Write(p []byte) (n int, err error) {
	cw.lg.Error(conv.BytesToStr(p), "OTEL", true)
	return len(p), nil
}

func OTelInit(ctx context.Context, conf *Config) (*trace.TracerProvider, error) {
	NewCustomWriter(customLogger.CopyLogger()).RedirectLoggerToSlog()

	exporter, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpointURL(conf.OTel.Endpoint))
	if err != nil {
		return nil, fmt.Errorf("init otel: %w", err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(otelServiceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(res),
	)

	otel.SetTextMapPropagator(propagation.TraceContext{})
	otel.SetTracerProvider(tp)

	return tp, nil
}
