package logger

import (
	"encoding/json"
	"fmt"
	"time"
)

type LogEvent struct {
	Event     string `json:"event"`
	JobID     string `json:"job_id,omitempty"`
	Language  string `json:"language,omitempty"`
	Status    string `json:"status,omitempty"`
	Error     string `json:"error,omitempty"`
	Timestamp string `json:"timestamp"`
}

func Log(event LogEvent) {
	event.Timestamp = time.Now().Format(time.RFC3339)

	logJSON, _ := json.Marshal(event)

	fmt.Println(string(logJSON))
}
