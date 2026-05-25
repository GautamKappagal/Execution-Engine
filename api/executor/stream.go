package executor

import (
	"bytes"
	"sync"
)

type callbackWriter struct {
	onOutput func(string)
}

func (w callbackWriter) Write(p []byte) (int, error) {
	if w.onOutput != nil && len(p) > 0 {
		w.onOutput(string(p))
	}
	return len(p), nil
}

type lockedWriter struct {
	mu       sync.Mutex
	buffer   *bytes.Buffer
	callback callbackWriter
}

func newLockedWriter(buffer *bytes.Buffer, onOutput func(string)) *lockedWriter {
	return &lockedWriter{
		buffer:   buffer,
		callback: callbackWriter{onOutput: onOutput},
	}
}

func (w *lockedWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.buffer != nil {
		_, _ = w.buffer.Write(p)
	}
	_, _ = w.callback.Write(p)
	return len(p), nil
}
