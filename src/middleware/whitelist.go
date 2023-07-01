package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

var IPWhitelist = map[string]bool{
	"3.235.211.187":  true,
	"52.70.18.120":   true,
	"107.23.255.128": true,
	"107.23.255.129": true,
	"107.23.255.131": true,
	"107.23.255.132": true,
	"107.23.255.133": true,
	"107.23.255.134": true,
	"107.23.255.135": true,
	"107.23.255.137": true,
	"107.23.255.138": true,
	"107.23.255.139": true,
	"107.23.255.140": true,
	"107.23.255.141": true,
	"107.23.255.142": true,
	"107.23.255.143": true,
	"107.23.255.144": true,
	"107.23.255.145": true,
	"107.23.255.146": true,
	"107.23.255.147": true,
	"107.23.255.148": true,
	"107.23.255.149": true,
	"107.23.255.150": true,
	"107.23.255.151": true,
	"107.23.255.152": true,
	"107.23.255.153": true,
	"107.23.255.154": true,
	"107.23.255.155": true,
	"107.23.255.156": true,
	"107.23.255.157": true,
	"107.23.255.158": true,
	"107.23.255.159": true,

	"35.162.152.183": true,
	"52.38.28.241":   true,
	"52.35.67.149":   true,
	"54.149.215.237": true,
	"13.127.146.34":  true,
	"13.127.207.241": true,
	"13.232.235.243": true,
	"13.233.81.143":  true,

	"13.112.233.15":  true,
	"18.182.156.77":  true,
	"52.194.200.157": true,
	"54.250.57.56":   true,

	"3.64.99.234":   true,
	"3.69.80.51":    true,
	"3.120.160.95":  true,
	"3.121.144.151": true,
	"18.156.144.73": true,
	"18.184.214.33": true,
	"18.197.117.10": true,

	"13.54.58.4":     true,
	"13.238.1.253":   true,
	"13.239.156.114": true,
	"54.153.234.158": true,

	"18.228.69.72":   true,
	"18.228.167.221": true,
	"18.228.209.157": true,
	"18.228.209.53":  true,

	"3.0.35.31":     true,
	"3.1.111.112":   true,
	"13.228.169.5":  true,
	"52.220.50.179": true,

	"34.250.225.89": true,
	"52.30.208.221": true,
	"63.34.177.151": true,
	"63.35.2.11":    true,

	"3.96.250.82":  true,
	"3.97.68.46":   true,
	"52.60.203.46": true,
}

func IPWhiteListMiddleware(whitelist map[string]bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		if !whitelist[ip] {
			c.IndentedJSON(http.StatusForbidden, gin.H{
				"message": "You are not authorised to use this endpoint",
			})
			return
		}
	}
}