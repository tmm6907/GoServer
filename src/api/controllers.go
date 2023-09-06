package api

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

	scoreRoutes := router.Group("/scores")
	scoreRoutes.GET("/", h.GetScores)

	detailRoutes := router.Group("/details")
	detailRoutes.GET("/:id", h.GetScoreDetails)
}
