
# OpenTelemetry demo project

## Overview
This project includes two main components:
1. **Backend**: An in-memory key-value database with a REST API interface.
2. **API**: A REST API that interacts with the backend.

The configuration for both components can be controlled via environment variables. This document outlines the minimal and full configuration options for each component, along with setup details.

---

## Usage Flow

This section demonstrates the typical usage flow of the API for interacting with the backend's in-memory key-value database. The flow includes retrieving, storing, and deleting key-value pairs.


### Step 1: Retrieve a Non-Existent Key (`GET`)
Attempting to retrieve a key that does not exist in the database will result in a `not found` response.

**Request:**
```bash
curl --location 'http://localhost:8080/storage/testKey' \
--data ''
```

**Response:**
```json
"not found"
```

### Step 2: Store a Key-Value Pair (`POST`)
Store a new key-value pair in the database using the `POST` method.

**Request:**
```bash
curl --location 'http://localhost:8080/storage' \
--data '{
    "key":"testKey",
    "value":"testValue"
}'
```

**Response:**
```json
"ok"
```

### Step 3: Retrieve the Stored Key-Value Pair (`GET`)
Retrieve the previously stored key-value pair.

**Request:**
```bash
curl --location 'http://localhost:8080/storage/testKey' \
--data ''
```

**Response:**
```json
{
    "key": "testKey",
    "value": "testValue"
}
```

### Step 4: Delete the Key-Value Pair (`DELETE`)
Delete the stored key-value pair using the `DELETE` method.

**Request:**
```bash
curl --location --request DELETE 'http://localhost:8080/storage/testKey' \
--data ''
```

**Response:**
```json
"ok"
```

### Step 5: Attempt to Retrieve the Deleted Key (`GET`)
Attempting to retrieve a key that has been deleted will again result in a `not found` response.

**Request:**
```bash
curl --location 'http://localhost:8080/storage/testKey' \
--data ''
```

**Response:**
```json
"not found"
```

---

## API Configuration

### Environment Variables (Minimal Configuration)

| Variable            | Default Value         | Description                                |
|---------------------|-----------------------|--------------------------------------------|
| `BACKEND_ENDPOINT`  | `http://0.0.0.0:8081/storage` | URL of the backend storage API.            |
| `HTTP_HOST`         | `0.0.0.0`            | Host for the API HTTP server.             |
| `HTTP_PORT`         | `8080`               | Port for the API HTTP server.             |
| `OTEL_ENDPOINT`     | `http://localhost:4318` | OpenTelemetry OTelEndpoint for distributed tracing. |

### Full Configuration Options

#### General options
General configuration options can be found [here](https://github.com/KennyMacCormik/HerdMaster/blob/main/pkg/cfg/genCfg/gencfg.go).

#### Backend HTTP Client Configuration

##### Fields

- **BackendEndpoint**:
    - Description: Specifies the URL of the backend OTelEndpoint (e.g., REST API OTelEndpoint).
    - Validation: Must be a valid URL (`url` tag) and is required.
    - Example: `http://localhost:8081/storage`.

- **BackendRequestTimeout**:
    - Description: Specifies the maximum duration to wait for a backend request to complete.
    - Validation: Must be a duration between 100 ms and 30 s (`min=100ms, max=30s`).
    - Example: `5s`.

##### Usage


Example configuration for environment variables:

```env
BACKEND_ENDPOINT=http://localhost:8081/storage
BACKEND_REQUEST_TIMEOUT=5s
```

---

## Backend Configuration

### Environment Variables (Minimal Configuration)

| Variable        | Default Value         | Description                                |
|-----------------|-----------------------|--------------------------------------------|
| `HTTP_HOST`     | `0.0.0.0`            | Host for the backend HTTP server.         |
| `HTTP_PORT`     | `8081`               | Port for the backend HTTP server.         |
| `OTEL_ENDPOINT` | `http://localhost:4318` | OpenTelemetry OTelEndpoint for distributed tracing. |

#### General options
General configuration options can be found [here](https://github.com/KennyMacCormik/HerdMaster/blob/main/pkg/cfg/genCfg/gencfg.go).

---

## Key Features

1. **Backend**:
    - In-memory key-value database with RESTful API support.
    - Configurable via environment variables or full configuration.

2. **API**:
    - REST API that communicates with the backend using HTTP.
    - Integrated with OpenTelemetry for distributed tracing.

---

## TODO

- Add Swagger/OpenAPI documentation for the API and Backend endpoints.
- Define app behavior in more details
- Add prometheus exporter
- Fix main.go to remove useless comments and add emtpy strings
- Sort imports (Golnad have settings for that)
- Add linters (https://golangci-lint.run/usage/linters/)
- Change myinit name
- Change module name to repo name
- Add run and close funcs to http server
- Remove goroutine and closer return from InitServer()
- Change cfg package according to recommendations (https://github.com/katyafirstova/auth_service/tree/week_2)
- Change file naming to snake_case
- Change int to int64
- Move storage.go to separate folder
- Add empty strings everywhere
- Make logger global (refactor log package)
- Review http response codes (201, 204)
- Add structs and interfaces for handlers
- Do not use "Interface" and "layer" as names
- Add TTL tests for cache
- Fix compute.Get()
  - Remove unnecessary else
  - write val to cache
  - do not stop if cache fails
- use "service" instead of "compute"
- add common data types to models
- RateLimiter
  - Store RateLimiter conf in Redis
  - Move RateLimiter metrics to prometheus
  - Move RateLimiter away from logs and traces
  - Fix RateLimiter metrics
- Remove traces from middleware
- Add error marks to span

---

## License
This project is licensed under the [MIT License](https://opensource.org/licenses/MIT).

---

## Acknowledgments
Special thanks to the maintainers and contributors of the following libraries:
- [Gin Web Framework](https://github.com/gin-gonic/gin)
- [go-playground/validator](https://github.com/go-playground/validator)
- [OpenTelemetry Go](https://github.com/open-telemetry/opentelemetry-go)
