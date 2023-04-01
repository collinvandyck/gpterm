package log

import (
	"io"
	"os"
	"sync"
)

func FileWriter(path string) (io.WriteCloser, error) {
	var f *os.File
	var err error
	if path != "" {
		f, err = os.Create(path)
		if err != nil {
			return nil, err
		}
	}
	fw := &fileWriter{
		file:    f,
		maxSize: 1024 * 1024,
	}
	return fw, nil
}

type fileWriter struct {
	file    *os.File
	maxSize int64
	mut     sync.Mutex
}

func (w *fileWriter) Write(p []byte) (n int, err error) {
	w.mut.Lock()
	defer w.mut.Unlock()
	if w.file == nil {
		return 0, nil
	}
	n, err = w.file.Write(p)
	if err != nil {
		return
	}
	info, err := os.Stat(w.file.Name())
	if err != nil {
		return
	}
	if info.Size() > w.maxSize {
		err = w.file.Close()
		if err != nil {
			return
		}
		w.file, err = os.Create(w.file.Name())
		if err != nil {
			return
		}
	}
	return
}

func (w *fileWriter) Close() error {
	w.mut.Lock()
	defer w.mut.Unlock()
	if w.file == nil {
		return nil
	}
	err := w.file.Close()
	return err
}
