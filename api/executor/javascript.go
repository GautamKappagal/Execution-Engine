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

type JavaScriptExecutor struct{}

func (j JavaScriptExecutor) Execute(code string, input string) (string, error) {
	return j.ExecuteWithOutput(code, input, nil)
}

func (j JavaScriptExecutor) ExecuteWithOutput(code string, input string, onOutput func(chunk string)) (string, error) {
	// Create isolated execution directory
	jobDir, err := os.MkdirTemp("", "execution-*")
	if err != nil {
		return "", err
	}

	// Cleanup execution directory
	defer os.RemoveAll(jobDir)

	// Create main.js inside isolated directory
	filePath := filepath.Join(jobDir, "main.js")

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
		"-v",
		filePath+":/app/main.js:ro",
		"execution-javascript",
		"node",
		"/app/main.js",
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

	// Send input
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

	// Return stderr/stdout
	if err != nil {
		return outputBuf.String(), err
	}

	return outputBuf.String(), nil
}
