package telemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

type TracingOptions struct {
	Enabled     bool
	ServiceName string
	Version     string
	Env         string
	Endpoint    string // host:porta OTLP/HTTP (ex.: jaeger:4318)
	SampleRatio float64
}

// SetupTracing configura o TracerProvider global. Quando desabilitado, mantém
// o provider no-op do OpenTelemetry (custo praticamente zero) mas ainda
// registra o propagador W3C, para repassar traceparent recebido.
func SetupTracing(ctx context.Context, opts TracingOptions) (shutdown func(context.Context) error, err error) {
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{},
	))

	if !opts.Enabled {
		return func(context.Context) error { return nil }, nil
	}

	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(opts.Endpoint),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("criando exportador OTLP: %w", err)
	}

	// Sem WithSchemaURL de propósito: fixar a versão do semconv conflita com o
	// schema embutido no SDK (resource.Default) sempre que o SDK é atualizado.
	res, err := resource.New(ctx,
		resource.WithTelemetrySDK(),
		resource.WithHost(),
		resource.WithFromEnv(), // OTEL_RESOURCE_ATTRIBUTES
		resource.WithAttributes(
			semconv.ServiceName(opts.ServiceName),
			semconv.ServiceVersion(opts.Version),
			semconv.DeploymentEnvironment(opts.Env),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("criando resource OTel: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(opts.SampleRatio))),
	)
	otel.SetTracerProvider(tp)
	return tp.Shutdown, nil
}
