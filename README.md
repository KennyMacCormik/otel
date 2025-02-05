
# OpenTelemetry demo project

## Overview
This project includes two main components:
1. [**Backend**](https://github.com/KennyMacCormik/otel/tree/main/backend)
2. **API**: A REST API that interacts with the backend.

---

## Storage API Usage Flow

This document outlines the expected behavior of the storage API when interacting with keys that require URL encoding.

---

## Request Flow

### **1. Attempt to Retrieve a Non-Existent URL-Encoded Key**

#### Request:
```sh
curl --location 'http://localhost:8081/storage/te%24tKey' \
--data ''
```

#### Response:
- **Status Code:** `404 Not Found`

---

### **2. Attempt to Retrieve a Non-Existent Key that Doesn’t Require Encoding**

#### Request:
```sh
curl --location 'http://localhost:8081/storage/testKey' \
--data ''
```

#### Response:
- **Status Code:** `404 Not Found`

---

### **3. Attempt to Retrieve a Non-Encoded Key (Rejected)**

#### Request:
```sh
curl --location 'http://localhost:8081/storage/te$tKey' \
--data ''
```

#### Response:
- **Status Code:** `400 Bad Request`
- **Response Body:**
```json
{
    "error": "key must be URL-encoded"
}
```

---

### **4. Store a New Key-Value Pair (First Insertion)**

#### Request:
```sh
curl --location --request PUT 'http://localhost:8081/storage' \
--data '{
    "key":"te$tKey",
    "value":"testValue"
}'
```

#### Response:
- **Status Code:** `201 Created`

---

### **5. Store the Same Key-Value Pair (No Change)**

#### Request:
```sh
curl --location --request PUT 'http://localhost:8081/storage' \
--data '{
    "key":"te$tKey",
    "value":"testValue"
}'
```

#### Response:
- **Status Code:** `204 No Content`

---

### **6. Update Existing Key with a New Value**

#### Request:
```sh
curl --location --request PUT 'http://localhost:8081/storage' \
--data '{
    "key":"te$tKey",
    "value":"testValue1"
}'
```

#### Response:
- **Status Code:** `200 OK`

---

### **7. Retrieve the Stored Key-Value Pair**

#### Request:
```sh
curl --location 'http://localhost:8081/storage/te%24tKey' \
--data ''
```

#### Response:
- **Status Code:** `200 OK`
- **Response Body:**
```json
{
    "key": "te$tKey",
    "value": "testValue1"
}
```

---

### **8. Delete the Stored Key-Value Pair**

#### Request:
```sh
curl --location --request DELETE 'http://localhost:8081/storage/te%24tKey' \
--data ''
```

#### Response:
- **Status Code:** `204 No Content`

---

### **9. Verify Deletion of Key**

#### Request:
```sh
curl --location 'http://localhost:8081/storage/te%24tKey' \
--data ''
```

#### Response:
- **Status Code:** `404 Not Found`

---

## TODO

- Add Swagger/OpenAPI documentation for the API and Backend endpoints (api, ~~backend~~)
- Define app behavior in more details (api, ~~backend~~)
- Add prometheus exporter
- Fix main.go (and all other files too) to remove useless comments and add emtpy strings (api, backend)
- Sort imports (Golnad have settings for that) (api, ~~backend~~)
- Add [linters](https://golangci-lint.run/usage/linters/)
- Change myinit name (api, ~~backend~~, otel-common)
- Change module name (`module backend`) to repo name (`module github.com/KennyMacCormik/otel/backend`) (api, ~~backend~~)
- Add run and close funcs to http server (api, ~~backend~~)
- Remove goroutine and closer return from InitServer() (api, ~~backend~~)
- Change cfg package according to recommendations (https://github.com/katyafirstova/auth_service/tree/week_2) (api, ~~backend~~)
- Change file naming to snake_case (api, ~~backend~~)
- Change int to int64 (api, ~~backend~~)
- Move storage.go to separate folder (api, ~~backend~~)
- Add empty strings everywhere (api, ~~backend~~)
- Make logger global (refactor log package) (api, ~~backend~~)
- Review http response codes (201, 204) (api, ~~backend~~)
- Add structs and interfaces for handlers (api, ~~backend~~)
- Do not use "Interface" and "layer" as names (api, ~~backend~~)
- Add TTL tests for cache
- Fix compute.Get()
  - Remove unnecessary else
  - write val to cache
  - do not stop if cache fails
- use "service" instead of "compute"
- add common data types to models (api, ~~backend~~)
- RateLimiter
  - Store RateLimiter conf in Redis
  - Move RateLimiter metrics to prometheus
  - Move RateLimiter away from logs and traces
  - Fix RateLimiter metrics
- Remove traces from middleware (api, ~~backend~~)
- Add error marks to span (api, ~~backend~~)
- Add live and ready probes
- Add resource limits

## To ask

- Excessive error checks for request id in storage endpoint

---

## License
This project is licensed under the [MIT License](https://opensource.org/licenses/MIT).

---

## Acknowledgments
Special thanks to the maintainers and contributors of the following libraries:
- [Gin Web Framework](https://github.com/gin-gonic/gin)
- [go-playground/validator](https://github.com/go-playground/validator)
- [OpenTelemetry Go](https://github.com/open-telemetry/opentelemetry-go)
