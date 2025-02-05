
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

- Define app behavior in more details (api, ~~backend~~)
- Add prometheus exporter
- Add [linters](https://golangci-lint.run/usage/linters/)
- Review http response codes (201, 204) (api, ~~backend~~)
- Add interfaces for handlers (api, backend)
- Add TTL tests for cache
- Fix service.Get()
  - write value to cache
- RateLimiter ([example](https://github.com/uber-go/ratelimit))
  - Store RateLimiter conf in Redis
  - Move RateLimiter metrics to prometheus
  - Move RateLimiter away from logs and traces
  - Fix RateLimiter metrics
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
