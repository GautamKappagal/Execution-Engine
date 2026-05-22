package executor

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"time"
)

func ExecuteJavaScript(code string, input string) (string, error) {
	// Create temporary JavaScript file
	tempFile, err := os.CreateTemp("", "code-*.js")
	if err != nil {
		return "", err
	}

	// Cleanup temp file
	defer os.Remove(tempFile.Name())

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
		"--rm",
		"--network=none",
		"--memory=128m",
		"--cpus=0.5",
		"-v",
		tempFile.Name()+":/app/main.js",
		"execution-javascript",
		"node",
		"/app/main.js",
	)

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
	output, err := cmd.CombinedOutput()

	// Handle timeout
	if ctx.Err() == context.DeadlineExceeded {
		return "", errors.New("execution timed out")
	}

	// Return stderr/stdout
	if err != nil {
		return string(output), err
	}

	return string(output), nil
}
