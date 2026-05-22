package models

type ExecutionResult struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Output string `json:"output"`
	Error  string `json:"error"`
}
