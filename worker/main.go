package main

import (
	"context"
	"encoding/json"
	"fmt"

	"execution-engine-api/executor"
	"execution-engine-api/models"

	goredis "github.com/redis/go-redis/v9"
)

var ctx = context.Background()

func main() {
	client := goredis.NewClient(&goredis.Options{
		Addr: "localhost:6379",
	})

	fmt.Println("Worker started. Waiting for jobs...")

	for {
		result, err := client.BLPop(ctx, 0, "execution_queue").Result()
		if err != nil {
			fmt.Println("Redis error:", err)
			continue
		}

		jobJSON := result[1]

		var job models.ExecutionJob

		err = json.Unmarshal([]byte(jobJSON), &job)
		if err != nil {
			fmt.Println("Failed to parse job:", err)
			continue
		}

		fmt.Println("Executing job:", job.ID)

		runningResult := models.ExecutionResult{
			ID:     job.ID,
			Status: "running",
			Output: "",
			Error:  "",
		}

		runningJSON, _ := json.Marshal(runningResult)

		client.Set(
			ctx,
			"result:"+job.ID,
			runningJSON,
			0,
		)

		exec, exists := executor.Executors[job.Language]
		if !exists {
			fmt.Println("Unsupported language:", job.Language)
			continue
		}

		output, err := exec.Execute(job.Code, job.Input)
		if err != nil {
			fmt.Println("Execution error:", err)

			failedResult := models.ExecutionResult{
				ID:     job.ID,
				Status: "failed",
				Output: output,
				Error:  err.Error(),
			}

			failedJSON, _ := json.Marshal(failedResult)

			client.Set(
				ctx,
				"result:"+job.ID,
				failedJSON,
				0,
			)

			continue
		}

		fmt.Println("Execution output:")
		fmt.Println(output)

		completedResult := models.ExecutionResult{
			ID:     job.ID,
			Status: "completed",
			Output: output,
			Error:  "",
		}

		completedJSON, _ := json.Marshal(completedResult)

		client.Set(
			ctx,
			"result:"+job.ID,
			completedJSON,
			0,
		)
	}
}
