# Configuration Guide

This section outlines available environment variables to configure the backend server. These variables allow fine-tuned control over server behavior, logging, tracing, and rate-limiting mechanisms.

## Logging Configuration

| Environment Variable | Description                                                                                 |
|----------------------|---------------------------------------------------------------------------------------------|
| `LOG_FORMAT`         | The log format. Must be either `text` or `json`. Default value is `json`.                   |
| `LOG_LEVEL`          | The log level. Must be one of `debug`, `info`, `warn`, or `error`. Default value is `warn`. |

## HTTP Server Configuration

| Environment Variable        | Description                                                                                                                                 |
|-----------------------------|---------------------------------------------------------------------------------------------------------------------------------------------|
| `HTTP_HOST`                 | **Required parameter.** The IP address or hostname of the HTTP server. Must be a valid IPv4 address or RFC1123-compliant hostname.          |
| `HTTP_PORT`                 | **Required parameter.** The port number for the HTTP server. Must be between 1025 and 65535.                                                |
| `HTTP_READ_TIMEOUT`         | Maximum duration for reading an entire request, including the body. Must be between 100ms and 1s. Default value is `100ms`.                 |
| `HTTP_WRITE_TIMEOUT`        | Maximum duration before timing out a write of the response. Must be between 100ms and 1s. Default value is `100ms`.                         |
| `HTTP_IDLE_TIMEOUT`         | Maximum time to wait for the next request when keep-alive enabled. Must be between 100ms and 1s. Default value is `100ms`.                  |
| `HTTP_SHUTDOWN_TIMEOUT`     | Maximum duration to wait for active connections to close gracefully during shutdown. Must be between 100ms and 30s. Default value is `10s`. |

## Gin router Configuration

| Environment Variable    | Description                                                                                                          |
|-------------------------|----------------------------------------------------------------------------------------------------------------------|
| `GIN_MODE`              | Defines the mode in which Gin runs. Possible values: `debug`, `release`, or `test`. The default value is `release`.  | 

## OpenTelemetry (OTel) Tracing Configuration

| Environment Variable         | Description                                                                                                                                         |
|------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------|
| `OTEL_ENDPOINT`              | **Required parameter.** The URL of the OTel exporter endpoint (e.g., OTLP HTTP/JSON). Must be a valid URL. Does not support `https://` endpoints.   |
| `OTEL_SHUTDOWN_TIMEOUT`      | Maximum duration to wait for graceful shutdown of the tracing system. Must be between 100ms and 30s.  Default value is `500ms`                      |

## Rate Limiter Configuration

| Environment Variable          | Description                                                                                                                                        |
|-------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------|
| `RATE_LIMITER_MAX_CONN`       | Maximum number of concurrent requests allowed. Must be between 1 and 100,000. Default value is `100`.                                              |
| `RATE_LIMITER_MAX_WAIT`       | Maximum number of requests allowed to wait when the limit is reached. Must be between 1 and 100,000. Default value is `100`.                       |
| `RATE_LIMITER_RETRY_AFTER`    | The `Retry-After` header value in seconds when a request is rejected due to rate limiting. Must be between 1 and 60 seconds. Default value is `1`. |

## Usage

Set the environment variables before running the service. For example:

```sh
export GRPC_HOST=127.0.0.1
export GRPC_PORT=50051
export LOG_FORMAT=json
export LOG_LEVEL=info
export HTTP_HOST=0.0.0.0
export HTTP_PORT=8080
export OTEL_ENDPOINT="http://otel-collector.example.com:4318"
export RATE_LIMITER_MAX_CONN=1000
```

These parameters ensure proper configuration of the microservices, enabling secure, efficient, and scalable operation.

