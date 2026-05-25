package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type executeRequest struct {
	Language string `json:"language"`
	Code     string `json:"code"`
	Input    string `json:"input"`
}

type executeResponse struct {
	JobID   string `json:"job_id"`
	Message string `json:"message"`
	Error   string `json:"error"`
}

type executionResult struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Output string `json:"output"`
	Error  string `json:"error"`
}

func main() {
	var apiBase string
	var language string
	var filePath string
	var codeInline string
	var inputPath string
	var wait bool
	var stream bool
	var pollInterval time.Duration
	var timeout time.Duration

	flag.StringVar(&apiBase, "api", envOrDefault("EXEC_API_URL", "http://localhost:8080"), "API base URL")
	flag.StringVar(&language, "lang", "", "Language (python|javascript|cpp)")
	flag.StringVar(&filePath, "file", "", "Code file path (use '-' for stdin)")
	flag.StringVar(&codeInline, "code", "", "Inline code string (alternative to --file)")
	flag.StringVar(&inputPath, "input", "", "Input file path (optional; use '-' for stdin)")
	flag.BoolVar(&wait, "wait", true, "Wait for completion by polling /result/:id")
	flag.BoolVar(&stream, "stream", false, "Stream logs via WebSocket /ws/:id while waiting")
	flag.DurationVar(&pollInterval, "poll", 250*time.Millisecond, "Polling interval (when --wait is true)")
	flag.DurationVar(&timeout, "timeout", 15*time.Second, "Overall CLI timeout (0 for no timeout)")
	flag.Parse()

	if strings.TrimSpace(language) == "" {
		fatalf("missing --lang")
	}

	code, err := readCode(filePath, codeInline)
	if err != nil {
		fatalf("read code: %v", err)
	}

	input, err := readOptionalInput(inputPath)
	if err != nil {
		fatalf("read input: %v", err)
	}

	ctx := context.Background()
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	jobID, err := submit(ctx, apiBase, executeRequest{
		Language: language,
		Code:     code,
		Input:    input,
	})
	if err != nil {
		fatalf("submit: %v", err)
	}

	fmt.Printf("job_id=%s\n", jobID)

	if !wait {
		return
	}

	var streamDone chan struct{}
	if stream {
		streamDone = make(chan struct{})
		go func() {
			defer close(streamDone)
			_ = streamLogs(ctx, apiBase, jobID, os.Stdout)
		}()
	}

	result, err := waitForResult(ctx, apiBase, jobID, pollInterval)
	if err != nil {
		fatalf("wait: %v", err)
	}

	if streamDone != nil {
		select {
		case <-streamDone:
		default:
		}
	}

	fmt.Printf("status=%s\n", result.Status)
	if result.Output != "" {
		fmt.Print(result.Output)
	}
	if result.Error != "" {
		fmt.Fprintf(os.Stderr, "error=%s\n", result.Error)
	}
}

func submit(ctx context.Context, apiBase string, req executeRequest) (string, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(apiBase, "/")+"/execute", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return "", err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("http %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var out executeResponse
	if err := json.Unmarshal(respBody, &out); err != nil {
		return "", err
	}
	if out.JobID == "" {
		if out.Error != "" {
			return "", errors.New(out.Error)
		}
		return "", fmt.Errorf("missing job_id in response: %s", strings.TrimSpace(string(respBody)))
	}
	return out.JobID, nil
}

func waitForResult(ctx context.Context, apiBase string, jobID string, pollInterval time.Duration) (executionResult, error) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		result, found, err := fetchResult(ctx, apiBase, jobID)
		if err != nil {
			return executionResult{}, err
		}
		if found && (result.Status == "completed" || result.Status == "failed") {
			return result, nil
		}

		select {
		case <-ctx.Done():
			return executionResult{}, ctx.Err()
		case <-ticker.C:
		}
	}
}

func fetchResult(ctx context.Context, apiBase string, jobID string) (executionResult, bool, error) {
	url := strings.TrimRight(apiBase, "/") + "/result/" + jobID
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return executionResult{}, false, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return executionResult{}, false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return executionResult{}, false, nil
	}

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return executionResult{}, false, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return executionResult{}, false, fmt.Errorf("http %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var result executionResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		return executionResult{}, false, err
	}

	return result, true, nil
}

func streamLogs(ctx context.Context, apiBase string, jobID string, w io.Writer) error {
	wsURL := strings.TrimRight(apiBase, "/")
	wsURL = strings.TrimPrefix(wsURL, "http://")
	wsURL = strings.TrimPrefix(wsURL, "https://")

	scheme := "ws"
	if strings.HasPrefix(apiBase, "https://") {
		scheme = "wss"
	}

	url := fmt.Sprintf("%s://%s/ws/%s", scheme, wsURL, jobID)

	dialer := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}

	conn, _, err := dialer.DialContext(ctx, url, nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		_, msg, err := conn.ReadMessage()
		if err != nil {
			return err
		}

		if len(msg) > 0 {
			if _, err := w.Write(msg); err != nil {
				return err
			}
			if msg[len(msg)-1] != '\n' {
				if _, err := io.WriteString(w, "\n"); err != nil {
					return err
				}
			}
		}
	}
}

func readCode(filePath, codeInline string) (string, error) {
	if strings.TrimSpace(codeInline) != "" {
		return codeInline, nil
	}
	if strings.TrimSpace(filePath) == "" {
		return "", errors.New("provide --file or --code")
	}
	if filePath == "-" {
		b, err := io.ReadAll(io.LimitReader(os.Stdin, 1<<20))
		return string(b), err
	}
	b, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	if len(b) > 1<<20 {
		return "", errors.New("code file too large (max 1MB)")
	}
	return string(b), nil
}

func readOptionalInput(inputPath string) (string, error) {
	if strings.TrimSpace(inputPath) == "" {
		return "", nil
	}
	if inputPath == "-" {
		b, err := io.ReadAll(io.LimitReader(os.Stdin, 1<<20))
		return string(b), err
	}
	b, err := os.ReadFile(inputPath)
	if err != nil {
		return "", err
	}
	if len(b) > 1<<20 {
		return "", errors.New("input file too large (max 1MB)")
	}
	return string(b), nil
}

func envOrDefault(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(2)
}

