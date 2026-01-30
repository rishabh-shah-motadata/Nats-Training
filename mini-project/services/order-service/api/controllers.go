package api

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/nats-io/nats.go"
)

func RegisterRoutes(router *gin.Engine, pg *pgxpool.Pool, js nats.JetStreamContext) {
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	router.POST("/order", saveOrderHandler(pg, js))
}
