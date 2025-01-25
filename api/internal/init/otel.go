package init

import (
	"context"
	"fmt"
	"github.com/KennyMacCormik/HerdMaster/pkg/conv"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
	"log"
	"log/slog"
	"time"
)

const otelServiceName = "api"

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

type OtelConfig struct {
	Endpoint        string        `mapstructure:"otel_endpoint" validate:"url,required"`
	ShutdownTimeout time.Duration `mapstructure:"otel_shutdown_timeout" validate:"min=100ms,max=30s"`
}

func InitOtel(ctx context.Context, conf *Config, lg *slog.Logger) (closer func() error, err error) {
	NewCustomWriter(lg).RedirectLoggerToSlog()

	exporter, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpointURL(conf.Otel.Endpoint))
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

	return func() error {
		ctxStop, cancel := context.WithTimeout(ctx, conf.Otel.ShutdownTimeout)
		defer cancel()
		return tp.Shutdown(ctxStop)
	}, nil
}
