package middleware

import (
	"crypto/subtle"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

const HeaderAdmin = "X-Admin-Token"

func RequiereAdmin(clave string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		recibida := ctx.GetHeader(HeaderAdmin)
		// clave == "" significaría que cualquiera con cabecera vacía entra.
		if clave == "" || subtle.ConstantTimeCompare([]byte(recibida), []byte(clave)) != 1 {
			fmt.Printf("[REST] %s %s -> RECHAZADO: clave de administrador inválida\n", ctx.Request.Method, ctx.Request.URL.Path)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "clave de administrador inválida"})
			return
		}
		ctx.Next()
	}
}
