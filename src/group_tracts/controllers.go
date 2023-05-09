package group_tracts

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type handler struct {
	DB *gorm.DB
}

func RegisterRoutes(router *gin.Engine, db *gorm.DB) {
	h := &handler{
		DB: db,
	}

	routes := router.Group("/scores")
	routes.GET("/address", h.GetScoreByAddress)
	routes.GET("/", h.GetScores)
}
