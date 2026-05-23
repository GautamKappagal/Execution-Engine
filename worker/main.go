package main

import (
	"context"
	"encoding/json"
	"fmt"

	"execution-engine-api/executor"
	"execution-engine-api/logger"
	"execution-engine-api/models"

	goredis "github.com/redis/go-redis/v9"
)

var ctx = context.Background()

func main() {

	client := goredis.NewClient(&goredis.Options{
		Addr: "redis:6379",
	})

	fmt.Println("Worker started. Waiting for jobs...")

	for {

		result, err := client.BLPop(ctx, 0, "execution_queue").Result()
		if err != nil {

			logger.Log(logger.LogEvent{
				Event: "redis_error",
				Error: err.Error(),
			})

			continue
		}

		jobJSON := result[1]

		var job models.ExecutionJob

		err = json.Unmarshal([]byte(jobJSON), &job)
		if err != nil {

			logger.Log(logger.LogEvent{
				Event: "job_parse_failed",
				Error: err.Error(),
			})

			continue
		}

		// Job picked up by worker
		logger.Log(logger.LogEvent{
			Event:    "job_started",
			JobID:    job.ID,
			Language: job.Language,
			Status:   "running",
		})

		runningResult := models.ExecutionResult{
			ID:     job.ID,
			Status: "running",
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

			logger.Log(logger.LogEvent{
				Event:    "unsupported_language",
				JobID:    job.ID,
				Language: job.Language,
				Status:   "failed",
			})

			failedResult := models.ExecutionResult{
				ID:     job.ID,
				Status: "failed",
				Error:  "unsupported language",
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

		output, err := exec.Execute(job.Code, job.Input)

		if err != nil {

			status := "failed"

			if err.Error() == "execution timed out" {
				status = "timed_out"
			}

			logger.Log(logger.LogEvent{
				Event:    "job_failed",
				JobID:    job.ID,
				Language: job.Language,
				Status:   status,
				Error:    err.Error(),
			})

			failedResult := models.ExecutionResult{
				ID:     job.ID,
				Status: status,
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

		// Publish execution logs
		client.Publish(
			ctx,
			"logs:"+job.ID,
			output,
		)

		logger.Log(logger.LogEvent{
			Event:    "job_completed",
			JobID:    job.ID,
			Language: job.Language,
			Status:   "completed",
		})

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
