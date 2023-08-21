package middleware

import (
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"nwi.io/nwi/caches"
	"nwi.io/nwi/serializers"
)

func HandleCachedResults() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		url := ctx.Request.URL.RequestURI()
		cacheResult, ok := caches.CACHE.Get(url)
		if ok {
			switch t := reflect.TypeOf(cacheResult); t {
			case reflect.TypeOf([]serializers.ScoreResults{}), reflect.TypeOf(serializers.AddressScoreResult{}), reflect.TypeOf(serializers.DetailResult{}):
				ctx.JSON(http.StatusOK, &cacheResult)
				return
			case reflect.TypeOf(serializers.AddressScoreResultXML{}), reflect.TypeOf(serializers.XMLResults{}), reflect.TypeOf(serializers.DetailResultXML{}):
				ctx.XML(http.StatusOK, &cacheResult)
				return
			default:

			}
		}
		ctx.Next()
	}
}
