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

		exec, exists := executor.Executors[job.Language]
		if !exists {
			fmt.Println("Unsupported language:", job.Language)
			continue
		}

		output, err := exec.Execute(job.Code, job.Input)
		if err != nil {
			fmt.Println("Execution error:", err)
			fmt.Println("Output:", output)
			continue
		}

		fmt.Println("Execution output:")
		fmt.Println(output)
	}
}
