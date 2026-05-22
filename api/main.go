package main

import (
	"net/http"

	"execution-engine-api/executor"

	"github.com/gin-gonic/gin"
)

type ExecuteRequest struct {
	Code  string `json:"code"`
	Input string `json:"input"`
}

type ExecuteResponse struct {
	Output string `json:"output"`
}

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

		output, err := executor.ExecutePython(req.Code, req.Input)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":  err.Error(),
				"output": output,
			})
			return
		}

		c.JSON(http.StatusOK, ExecuteResponse{
			Output: output,
		})
	})

	r.Run(":8080")
}
