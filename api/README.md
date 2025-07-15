# Description

## Overview
The **API** is a RESTful interface for managing key-value storage, designed for efficient retrieval and modification of stored data. It acts as a middleware between clients and the [Backend Service](https://github.com/KennyMacCormik/otel/tree/main/backend), ensuring optimized data access and enhanced performance.

Key Features
- Simple RESTful Interface – Supports `PUT`, `GET`, and `DELETE` operations for managing key-value pairs.
- Integrated Caching – Built-in caching mechanism with a 60-second TTL, reducing backend load and improving response times.
- Backend Integration – Fetches and updates values from the [Backend Service](https://github.com/KennyMacCormik/otel/tree/main/backend), ensuring consistency.
- Optimized for Performance – Implements caching strategies to minimize redundant backend requests, improving throughput.
- Observability with OpenTelemetry – Tracing support for distributed request tracking

## API Endpoints

### **Retrieve a Value by Key**
- **GET** `/storage/{key}`
- **Description**: Fetches the stored value for a given key. Key _**must**_ be URL-encoded.
- **Responses**:
    - `200 OK`: Returns the key-value pair.
    - `400 Bad Request`: Malformed request.
    - `404 Not Found`: Key does not exist.
    - `500 Internal Server Error`: Unexpected server error.

### **Store or Update a Key-Value Pair**
- **PUT** `/storage`
- **Description**: Creates or updates a key-value pair.
- **Request Body**:
    - `key` (string, required, _**mustn't**_ be URL-encoded)
    - `value` (string, required, _**mustn't**_ be URL-encoded)
- **Responses**:
    - `200 OK`: Successfully updated existing value.
    - `201 OK`: Successfully created a new value.
    - `204 OK`: Successful request, nothing changed.
    - `400 Bad Request`: Malformed request.
    - `500 Internal Server Error`: Unexpected server error.

### **Delete a Key-Value Pair**
- **DELETE** `/storage/{key}`
- **Description**: Removes a key-value pair from storage. Key _**must**_ be URL-encoded.
- **Responses**:
    - `204 OK`: Successfully deleted.
    - `400 Bad Request`: Malformed request.
    - `500 Internal Server Error`: Unexpected server error.

# Build Guide

To build the application, specify the target OS, architecture, and output executable name, use:

```sh
CGO_ENABLED=0 [GOOS=<TARGET_OS>] [GOARCH=<TARGET_ARCH>] go build -o <OUTPUT_EXEC_NAME>
```

## Parameters

- `GOOS` – Specifies the target operating system (e.g., `linux`, `windows`, `darwin`). This parameter can be omitted.
- `GOARCH` – Specifies the target architecture (e.g., `amd64`, `arm64`). This parameter can be omitted.

## Example

Building with defaults:

```sh
CGO_ENABLED=0 go build -o backend
```

Building for Linux with AMD64 architecture:

```sh
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o backend
```

# Configuration Guide

This section outlines available environment variables to configure the backend server. These variables allow fine-tuned control over server behavior, logging, tracing, and rate-limiting mechanisms.

## Backend Configuration

| Environment Variable             | Description                                                                                                |
|----------------------------------|------------------------------------------------------------------------------------------------------------|
| `BACKEND_CLIENT_ENDPOINT`        | **Required parameter.** The URL of the backend service. Must be a valid URL. Shall include necessary path  |
| `BACKEND_CLIENT_REQUEST_TIMEOUT` | Maximum duration of a request to backend service. Must be between 100ms and 1s.  Default value is `200ms`. |

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
export LOG_LEVEL=info
export HTTP_HOST=0.0.0.0
export HTTP_PORT=8080
export OTEL_ENDPOINT="http://otel-collector.example.com:4318"
export BACKEND_CLIENT_ENDPOINT="http://backend-service.example.com:8081"
```

These parameters ensure proper configuration of the microservices, enabling secure, efficient, and scalable operation.
