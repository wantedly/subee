package safe

import (
	"io"
	"sync"
)

type writerWrapper struct {
	original io.Writer
	mu       sync.Mutex
}

func NewWriter(original io.Writer) io.Writer {
	return &writerWrapper{original: original}
}

func (w *writerWrapper) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.original.Write(p)
}
