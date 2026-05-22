package executor

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"time"
)

type CPPExecutor struct{}

func (c CPPExecutor) Execute(code string, input string) (string, error) {
	// Create temp cpp file
	tempFile, err := os.CreateTemp("", "code-*.cpp")
	if err != nil {
		return "", err
	}

	defer os.Remove(tempFile.Name())

	_, err = tempFile.WriteString(code)
	if err != nil {
		return "", err
	}

	tempFile.Close()

	// Create timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Docker compile + execute command
	cmd := exec.CommandContext(
		ctx,
		"docker",
		"run",
		"-i",
		"--rm",
		"--network=none",
		"--memory=256m",
		"--cpus=1",
		"-v",
		tempFile.Name()+":/app/main.cpp",
		"execution-cpp",
		"sh",
		"-c",
		"g++ /app/main.cpp -o /app/main && /app/main",
	)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", err
	}

	go func() {
		defer stdin.Close()
		stdin.Write([]byte(input))
	}()

	output, err := cmd.CombinedOutput()

	if ctx.Err() == context.DeadlineExceeded {
		return "", errors.New("execution timed out")
	}

	if err != nil {
		return string(output), err
	}

	return string(output), nil
}
