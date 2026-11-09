package middleware

import (
	"crypto/subtle"
	"net/http"

	"github.com/gin-gonic/gin"
)

const HeaderGateway = "X-Gateway-Token"

// FiltroGateway rechaza toda petición que no traiga el secreto del gateway
func FiltroGateway(secreto string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ctx.GetHeader(HeaderGateway)
		if token == "" || subtle.ConstantTimeCompare([]byte(token), []byte(secreto)) != 1 {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "acceso denegado: la petición debe pasar por el API Gateway",
			})
		}
	}
}
