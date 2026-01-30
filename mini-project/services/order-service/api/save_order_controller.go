package api

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/nats-io/nats.go"
)

type order struct {
	ID     string  `json:"id" binding:"required"`
	Item   string  `json:"item" binding:"required"`
	Amount float64 `json:"amount" binding:"required"`
	Status string  `json:"status"`
}

func saveOrderHandler(pgxPool *pgxpool.Pool, js nats.JetStreamContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		var newOrder order
		if err := c.ShouldBindJSON(&newOrder); err != nil {
			log.Println("error binding order request payload", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
			return
		}
		newOrder.Status = "PENDING"

		// Start a transaction
		tx, err := pgxPool.Begin(ctx)
		if err != nil {
			log.Println("error starting transaction", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start transaction"})
			return
		}
		defer tx.Rollback(ctx)

		// Save order to PostgreSQL
		_, err = tx.Exec(ctx, "INSERT INTO orders (id, item, amount, status) VALUES ($1, $2, $3, $4)",
			newOrder.ID, newOrder.Item, newOrder.Amount, newOrder.Status)
		if err != nil {
			log.Println("error saving order to postgres", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save order"})
			return
		}

		// Publish order created event to NATS JetStream
		data := []byte(newOrder.ID)
		ack, err := js.Publish("orders.created", data)
		if err != nil {
			log.Println("error publishing order created event to NATS JetStream", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to publish order created event"})
			return
		}
		log.Println("published order created event with ack:", ack)

		// Commit the transaction
		err = tx.Commit(ctx)
		if err != nil {
			log.Println("error committing transaction", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to commit transaction"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"status": "order created", "order_id": newOrder.ID})
	}
}
