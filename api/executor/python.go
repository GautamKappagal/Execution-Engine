package executor

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type PythonExecutor struct{}

func (p PythonExecutor) Execute(code string, input string) (string, error) {
	return p.ExecuteWithOutput(code, input, nil)
}

func (p PythonExecutor) ExecuteWithOutput(code string, input string, onOutput func(chunk string)) (string, error) {
	// Create isolated execution directory
	jobDir, err := os.MkdirTemp("", "execution-*")
	if err != nil {
		return "", err
	}

	// Cleanup execution directory after execution
	defer os.RemoveAll(jobDir)

	// Create main.py inside isolated directory
	filePath := filepath.Join(jobDir, "main.py")

	tempFile, err := os.Create(filePath)
	if err != nil {
		return "", err
	}

	// Write code into file
	_, err = tempFile.WriteString(code)
	if err != nil {
		return "", err
	}

	tempFile.Close()

	// Create timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Docker execution command
	cmd := exec.CommandContext(
		ctx,
		"docker",
		"run",
		"-i",
		"--init",
		"--rm",
		"--network=none",
		"--memory=128m",
		"--cpus=0.5",
		"--ulimit",
		"nofile=1024:1024",
		"--ulimit",
		"nproc=64:64",
		"--read-only",
		"--tmpfs",
		"/tmp:rw,nosuid,nodev,uid=1000,gid=1000",
		"--user",
		"1000:1000",
		"--cap-drop=ALL",
		"--pids-limit=64",
		"--security-opt=no-new-privileges",
		"-e",
		"PYTHONDONTWRITEBYTECODE=1",
		"-v",
		filePath+":/app/main.py:ro",
		"execution-python",
		"python",
		"/app/main.py",
	)

	var outputBuf bytes.Buffer
	outWriter := newLockedWriter(&outputBuf, onOutput)
	cmd.Stdout = outWriter
	cmd.Stderr = outWriter

	// Create stdin pipe
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", err
	}

	// Send input to container
	go func() {
		defer stdin.Close()
		stdin.Write([]byte(input))
	}()

	// Execute command
	if err := cmd.Start(); err != nil {
		return "", err
	}
	err = cmd.Wait()

	// Handle timeout
	if ctx.Err() == context.DeadlineExceeded {
		return "", errors.New("execution timed out")
	}

	// Return stderr/stdout together
	if err != nil {
		return outputBuf.String(), err
	}

	return outputBuf.String(), nil
}
