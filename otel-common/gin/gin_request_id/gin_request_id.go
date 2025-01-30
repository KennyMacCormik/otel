package gin_request_id

import (
	"fmt"
	"github.com/KennyMacCormik/common/log"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"reflect"
)

const RequestIDKey = "X-Request-ID"

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(RequestIDKey)
		if requestID == "" {
			requestID = genUuid()
		}

		c.Set(RequestIDKey, requestID)
		c.Writer.Header().Set(RequestIDKey, requestID)

		c.Next()
	}
}

// GetRequestIDFromCtx is a helper function that extracts requestID from gin context.
func GetRequestIDFromCtx(c *gin.Context) (string, error) {
	var requestID string

	requestID = c.Writer.Header().Get(RequestIDKey)

	if requestID == "" {
		uuidFromCtx, ok := c.Get(RequestIDKey)
		if !ok {
			requestID = fallbackUuidHandler(c)

			log.Warn(
				fmt.Sprintf(
					"generated new requestID after missing %s in request context", RequestIDKey,
				),
				"requestID", requestID,
			)

			return requestID, nil
		}

		requestID, ok = uuidFromCtx.(string)
		if !ok {
			log.Error("invalid X-Request-Id type",
				"type", reflect.TypeOf(uuidFromCtx),
				"value", uuidFromCtx,
			)

			return "", fmt.Errorf(
				"invalid X-Request-Id type: got %T, value: %v",
				uuidFromCtx,
				uuidFromCtx,
			)
		}
	}

	return requestID, nil
}

func fallbackUuidHandler(c *gin.Context) string {
	requestID := genUuid()

	c.Set(RequestIDKey, requestID)
	c.Writer.Header().Set(RequestIDKey, requestID)

	return requestID
}

func genUuid() string {
	return uuid.New().String()
}
