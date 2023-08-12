package middleware

// func AuthenticateRequest() gin.HandlerFunc {
// 	return func(ctx *gin.Context) {
// 		userAuth := ctx.GetHeader("X-RapidAPI-Proxy-Secret")
// 		authKey := os.Getenv("X_RAPIDAPI_PROXY_SECRET")
// 		if userAuth == authKey {
// 			return
// 		}
// 		ctx.AbortWithStatus(http.StatusProxyAuthRequired)
// 	}
// }
