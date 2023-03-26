package main

import (
	"context"
	"flag"
	"fmt"
	"go.opentelemetry.io/otel/trace"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.18.0"
)

var exporter string = "otlp"

func main() {
	flag.StringVar(&exporter, "exporter", exporter, "exporter type: otlp or jaeger")
	flag.Parse()

	var spanExporter sdktrace.SpanExporter

	// setup tracing exporter
	grpcConnectionTimeout := 3 * time.Second
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(context.Background(), grpcConnectionTimeout)
	defer cancel()

	if exporter == "jaeger" {
		log.Printf("using jaeger exporter")
		// new jaeger exporter
		jaegerExporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint("http://localhost:14268/api/traces")))
		if err != nil {
			panic(err)
		}
		spanExporter = jaegerExporter
	} else {
		log.Printf("using otlp exporter")
		// new otlp exporter
		otlpExporter, err := otlptracegrpc.New(ctx)
		if err != nil {
			panic(err)
		}
		spanExporter = otlpExporter
	}

	attrs := resource.WithAttributes(semconv.ServiceName("demo-service"))
	res, err := resource.New(ctx, attrs)
	if err != nil {
		panic(err)
	}

	spanProcessor := sdktrace.NewSimpleSpanProcessor(spanExporter)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSpanProcessor(spanProcessor),
		sdktrace.WithResource(res),
	)
	// set global trace provider
	otel.SetTracerProvider(tp)
	defer tp.Shutdown(ctx)

	// you can use any propagator you want
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
