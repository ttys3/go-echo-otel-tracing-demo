package main

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel/trace"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.18.0"
)

func main() {
	// setup tracing exporter
	grpcConnectionTimeout := 3 * time.Second
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(context.Background(), grpcConnectionTimeout)
	defer cancel()

	spanExporter, err := otlptracegrpc.New(ctx)
	if err != nil {
		panic(err)
	}
	spanProcessor := sdktrace.NewSimpleSpanProcessor(spanExporter)

	attrs := resource.WithAttributes(semconv.ServiceName("demo-service"))
	res, _ := resource.New(ctx, attrs)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(1))),
		sdktrace.WithSpanProcessor(spanProcessor),
		sdktrace.WithResource(res),
	)
	// set global trace provider
	otel.SetTracerProvider(tp)
	defer tp.Shutdown(ctx)

	b3propagator := b3.New(b3.WithInjectEncoding(b3.B3MultipleHeader))
	propagator := propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}, b3propagator)
	// set global propagator
	otel.SetTextMapPropagator(propagator)

	e := echo.New()
	e.Use(otelecho.Middleware("demo-service"))
	e.GET("/", func(c echo.Context) error {
		traceid := TraceID(c.Request().Context())
		return c.String(http.StatusOK, fmt.Sprintf("traceid: %s", traceid))
	})
	e.Logger.Fatal(e.Start(":1323"))
}

func TraceID(ctx context.Context) string {
	if span := trace.SpanContextFromContext(ctx); span.HasTraceID() {
		return span.TraceID().String()
	}
	return ""
}
