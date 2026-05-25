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

type CPPExecutor struct{}

func (c CPPExecutor) Execute(code string, input string) (string, error) {
	return c.ExecuteWithOutput(code, input, nil)
}

func (c CPPExecutor) ExecuteWithOutput(code string, input string, onOutput func(chunk string)) (string, error) {
	// Create isolated execution directory
	jobDir, err := os.MkdirTemp("", "execution-*")
	if err != nil {
		return "", err
	}

	defer os.RemoveAll(jobDir)

	filePath := filepath.Join(jobDir, "main.cpp")

	tempFile, err := os.Create(filePath)
	if err != nil {
		return "", err
	}

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
		"--init",
		"--rm",
		"--network=none",
		"--memory=256m",
		"--cpus=1",
		"--ulimit",
		"nofile=1024:1024",
		"--ulimit",
		"nproc=128:128",
		"--read-only",
		"--tmpfs",
		"/tmp:rw,nosuid,nodev,uid=1000,gid=1000",
		"--tmpfs",
		"/work:rw,exec,nosuid,nodev,size=128m,uid=1000,gid=1000",
		"--workdir",
		"/work",
		"--user",
		"1000:1000",
		"--cap-drop=ALL",
		"--pids-limit=128",
		"--security-opt=no-new-privileges",
		"-v",
		filePath+":/src/main.cpp:ro",
		"execution-cpp",
		"sh",
		"-c",
		"g++ /src/main.cpp -O2 -pipe -o /work/main && /work/main",
	)

	var outputBuf bytes.Buffer
	outWriter := newLockedWriter(&outputBuf, onOutput)
	cmd.Stdout = outWriter
	cmd.Stderr = outWriter

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", err
	}

	go func() {
		defer stdin.Close()
		stdin.Write([]byte(input))
	}()

	if err := cmd.Start(); err != nil {
		return "", err
	}
	err = cmd.Wait()

	if ctx.Err() == context.DeadlineExceeded {
		return "", errors.New("execution timed out")
	}

	if err != nil {
		return outputBuf.String(), err
	}

	return outputBuf.String(), nil
}
