package router

import (
	"github.com/gin-gonic/gin"
)

// NewGinRouter create a new gin router for each micro service
func NewGinRouter() (router *gin.Engine) {
	router = gin.New()
	
	router.Use(gin.Recovery())

	return router
}
