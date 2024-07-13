package ioc

import (
	"context"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"time"
)

func InitOTEL() func(ctx context.Context) {
	type Config struct {
		ServiceName    string `yaml:"serviceName"`
		ServiceVersion string `yaml:"serviceVersion"`
	}
	var config Config
	err := viper.UnmarshalKey("opentelemetry", &config)

	res, err := newResource(config.ServiceName, config.ServiceVersion)
	if err != nil {
		panic(err)
	}
	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	tp, err := newTraceProvider(res)
	if err != nil {
		panic(err)
	}
	otel.SetTracerProvider(tp)
	return func(ctx context.Context) {
		tp.Shutdown(ctx)
	}
}

func newResource(serviceName, serviceVersion string) (*resource.Resource, error) {
	return resource.Merge(resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL,
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion)))
}

func newTraceProvider(res *resource.Resource) (*trace.TracerProvider, error) {
	type Config struct {
		CollectorURL string `yaml:"collectorURL"`
	}
	var config Config

	exporter, err := zipkin.New(
		config.CollectorURL)
	if err != nil {
		return nil, err
	}

	traceProvider := trace.NewTracerProvider(
		trace.WithBatcher(exporter,
			// Default is 5s. Set to 1s for demonstrative purposes.
			trace.WithBatchTimeout(time.Second)),
		trace.WithResource(res),
	)
	return traceProvider, nil
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{})
}