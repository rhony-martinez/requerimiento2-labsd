package middleware

// import (
// 	"net/http"

// 	"github.com/gin-gonic/gin"
// )

// CORS permite que el cliente HTML (servido desde otro origen o desde file://) consuma la API.
// func CORS() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		c.Header("Access-Control-Allow-Origin", "*")
// 		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
// 		c.Header("Access-Control-Allow-Headers", "Content-Type")
// 		if c.Request.Method == http.MethodOptions {
// 			c.AbortWithStatus(http.StatusNoContent)
// 			return
// 		}
// 		c.Next()
// 	}
// }
