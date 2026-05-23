package main

import (
	"encoding/json"
	"net/http"

	"execution-engine-api/executor"
	"execution-engine-api/models"
	"execution-engine-api/redis"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type ExecuteRequest struct {
	Language string `json:"language"`
	Code     string `json:"code"`
	Input    string `json:"input"`
}

type ExecuteResponse struct {
	Output string `json:"output"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

const (
	MaxCodeSize  = 10000
	MaxInputSize = 5000
)

func main() {
	r := gin.Default()

	r.POST("/execute", func(c *gin.Context) {
		var req ExecuteRequest

		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid request",
			})

			return
		}

		if len(req.Code) > MaxCodeSize {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "code exceeds maximum allowed size",
			})
			return
		}

		if len(req.Input) > MaxInputSize {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "input exceeds maximum allowed size",
			})
			return
		}

		_, exists := executor.Executors[req.Language]

		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "unsupported language",
			})

			return
		}

		job := models.ExecutionJob{
			ID:       uuid.New().String(),
			Language: req.Language,
			Code:     req.Code,
			Input:    req.Input,
		}

		jobJSON, err := json.Marshal(job)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to serialize job",
			})

			return
		}

		err = redis.Client.RPush(redis.Ctx, "execution_queue", jobJSON).Err()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to enqueue job",
			})

			return
		}

		result := models.ExecutionResult{
			ID:     job.ID,
			Status: "queued",
			Output: "",
			Error:  "",
		}

		resultJSON, err := json.Marshal(result)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to serialize result",
			})

			return
		}

		err = redis.Client.Set(
			redis.Ctx,
			"result:"+job.ID,
			resultJSON,
			0,
		).Err()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to store result",
			})

			return
		}

		c.JSON(http.StatusAccepted, gin.H{
			"message": "job queued",
			"job_id":  job.ID,
		})
	})

	r.GET("/result/:id", func(c *gin.Context) {
		id := c.Param("id")

		resultJSON, err := redis.Client.Get(
			redis.Ctx,
			"result:"+id,
		).Result()

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "result not found",
			})

			return
		}

		var result models.ExecutionResult

		err = json.Unmarshal([]byte(resultJSON), &result)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to parse result",
			})

			return
		}

		c.JSON(http.StatusOK, result)
	})

	r.GET("/ws/:id", func(c *gin.Context) {
		jobID := c.Param("id")

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to upgrade to WebSocket",
			})
			return
		}

		defer conn.Close()

		pubsub := redis.Client.Subscribe(
			redis.Ctx,
			"logs:"+jobID,
		)

		defer pubsub.Close()

		ch := pubsub.Channel()

		for msg := range ch {
			err := conn.WriteMessage(
				websocket.TextMessage,
				[]byte(msg.Payload),
			)

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "failed to send WebSocket message",
				})

				break
			}
		}
	})

	r.Run(":8080")
}
