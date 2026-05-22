package models

type ExecutionJob struct {
	ID       string `json:"id"`
	Language string `json:"language"`
	Code     string `json:"code"`
	Input    string `json:"input"`
}
