package executor

var Executors = map[string]Executor{
	"python":     PythonExecutor{},
	"javascript": JavaScriptExecutor{},
	"cpp":        CPPExecutor{},
}
