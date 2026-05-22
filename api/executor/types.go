package executor

type Executor interface {
	Execute(code string, input string) (string, error)
}
