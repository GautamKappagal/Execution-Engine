package executor

type Executor interface {
	Execute(code string, input string) (string, error)
}

// StreamingExecutor can emit execution output incrementally (best-effort).
// The final returned output should contain the full combined output.
type StreamingExecutor interface {
	ExecuteWithOutput(code string, input string, onOutput func(chunk string)) (string, error)
}
