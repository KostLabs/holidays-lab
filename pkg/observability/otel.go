package observability

import (
	"context"
	"log"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// InitProvider configures a global tracer provider using an OTLP/HTTP exporter.
// Endpoint, headers, and other options are taken from OTEL_EXPORTER_OTLP_* env vars.
func InitProvider(ctx context.Context, serviceName string) func(context.Context) error {
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	tracesEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT")
	protocol := os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL")
	compression := os.Getenv("OTEL_EXPORTER_OTLP_COMPRESSION")
	headersEnv := os.Getenv("OTEL_EXPORTER_OTLP_HEADERS")
	licenseKey := os.Getenv("NEW_RELIC_LICENSE_KEY")

	licenseSet := licenseKey != ""
	headersSet := headersEnv != ""

	log.Printf("otel: initializing provider for service=%s endpoint=%q tracesEndpoint=%q protocol=%q compression=%q headersEnvSet=%t licenseSet=%t", serviceName, endpoint, tracesEndpoint, protocol, compression, headersSet, licenseSet)

	var (
		exporter *otlptrace.Exporter
		err      error
	)

	if licenseSet {
		log.Printf("otel: using NEW_RELIC_LICENSE_KEY for api-key header (service=%s)", serviceName)
		exporter, err = otlptracehttp.New(ctx, otlptracehttp.WithHeaders(map[string]string{
			"api-key": licenseKey,
		}))
	} else {
		log.Printf("otel: NEW_RELIC_LICENSE_KEY not set, relying on OTEL_EXPORTER_OTLP_HEADERS (service=%s)", serviceName)
		exporter, err = otlptracehttp.New(ctx)
	}
	if err != nil {
		log.Printf("otel: failed to create OTLP exporter: %v", err)
		return func(context.Context) error { return nil }
	}

	hostname, _ := os.Hostname()
	instanceID := hostname + ":" + serviceName + ":" + time.Now().Format("20060102150405")
	serviceVersion := os.Getenv("SERVICE_VERSION")

	attrs := []attribute.KeyValue{
		semconv.ServiceNameKey.String(serviceName),
		semconv.ServiceInstanceIDKey.String(instanceID),
	}
	if serviceVersion != "" {
		attrs = append(attrs, semconv.ServiceVersionKey.String(serviceVersion))
	}

	res, err := resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
		resource.WithAttributes(attrs...),
	)
	if err != nil {
		log.Printf("otel: failed to create resource: %v", err)
		res = resource.Default()
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	log.Printf("otel: tracer provider created for service=%s", serviceName)

	// Log any async OTEL errors (e.g. export failures).
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		if err != nil {
			log.Printf("otel: async error for service=%s: %v", serviceName, err)
		}
	}))

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	log.Printf("otel: global tracer provider and propagator configured for service=%s", serviceName)

	return func(ctx context.Context) error {
		log.Printf("otel: shutting down provider for service=%s", serviceName)
		return tp.Shutdown(ctx)
	}
}
