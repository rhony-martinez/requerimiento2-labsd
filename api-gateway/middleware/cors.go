package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set("Access-Control-Allow-Origin", "*")
		h.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		h.Set("Access-Control-Allow-Headers",
			"Content-Type, X-Grpc-Web, X-User-Agent, Grpc-Timeout, Keep-Alive, Cache-Control, "+
				"X-Accept-Content-Transfer-Encoding, X-Accept-Content-Type, Content-Transfer-Encoding, "+
				"X-Usuario-Id, X-Audio-Id, Idempotency-Key")
		h.Set("Access-Control-Expose-Headers", "Grpc-Status, Grpc-Message")
		h.Set("Access-Control-Max-Age", "86400")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
