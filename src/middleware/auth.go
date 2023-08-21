package middleware

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func AuthenticateRequest() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		secretKey := os.Getenv("X_RAPIDAPI_PROXY_SECRET")
		secret := ctx.GetHeader("X-RapidAPI-Proxy-Secret")
		if secret == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "API key missing secret key"})
			return
		}
		if secret != secretKey {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key or secret key"})
			return
		}
		ctx.Next()
	}
}
