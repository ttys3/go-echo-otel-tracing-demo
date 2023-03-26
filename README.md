# go-echo-otel-tracing-demo

## how to run the demo?

### 1. setup tracing collector

#### otlp

see oltp collector docs https://github.com/open-telemetry/opentelemetry-collector-contrib

or tempo docs https://grafana.com/docs/tempo/latest/getting-started/

and then run grafana to see the tracing graph

    
#### jaeger

```bash
docker run -d --name jaeger \
  -e COLLECTOR_ZIPKIN_HTTP_PORT=9411 \
  -p 5775:5775/udp \
  -p 6831:6831/udp \
  -p 6832:6832/udp \
  -p 5778:5778 \
  -p 16686:16686 \
  -p 14268:14268 \
  -p 14250:14250 \
  -p 9411:9411 \
  jaegertracing/all-in-one:1.43
```

### 2. run the demo

```bash
go run .
```

### 3. see the tracing graph

```shell
./test.sh

# or
./test.sh remote
```