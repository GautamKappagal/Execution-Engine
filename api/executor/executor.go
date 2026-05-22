package executor

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"time"
)

func ExecutePython(code string, input string) (string, error) {
	// Create temporary Python file
	tempFile, err := os.CreateTemp("", "code-*.py")
	if err != nil {
		return "", err
	}

	// Cleanup temp file after execution
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
		tempFile.Name()+":/app/main.py",
		"execution-python",
		"python",
		"/app/main.py",
	)

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
	output, err := cmd.CombinedOutput()

	// Handle timeout
	if ctx.Err() == context.DeadlineExceeded {
		return "", errors.New("execution timed out")
	}

	// Return stderr/stdout together
	if err != nil {
		return string(output), err
	}

	return string(output), nil
}
