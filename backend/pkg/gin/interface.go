package gin

import "github.com/gin-gonic/gin"

type GinHandler interface {
	GetGinHandler() func(*gin.Engine)
}
