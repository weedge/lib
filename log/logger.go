package log

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

var _ io.WriteCloser = (*Logger)(nil)

var (
	// os_Stat exists so it can be mocked out by tests.
	os_Stat = os.Stat
)

type Logger struct {
	Filename string
	file     *os.File
	mu       sync.Mutex
}

// dir returns the directory for the current filename.
func (l *Logger) dir() string {
	return filepath.Dir(l.Filename)
}

// openNew opens a new log file for writing, moving any old log file out of the
// way.  This methods assumes the file has already been closed.
func (l *Logger) openNew() error {
	err := os.MkdirAll(l.dir(), 0755)
	if err != nil {
		return fmt.Errorf("can't make directories for new logfile: %s", err)
	}
	f, err := os.OpenFile(l.Filename, os.O_CREATE|os.O_WRONLY, 0655)
	if err != nil {
		return fmt.Errorf("can't open new logfile: %s", err)
	}
	l.file = f
	return nil
}

func (l *Logger) openExistingOrNew() error {
	filename := l.Filename
	if _, err := os_Stat(filename); os.IsNotExist(err) {
		return l.openNew()
	}

	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0655)
	l.file = file
	if err != nil {
		return l.openNew()
	}
	return nil
}

func (l *Logger) Write(p []byte) (n int, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file == nil {
		if err = l.openExistingOrNew(); err != nil {
			return 0, err
		}
	}

	n, err = l.file.Write(p)

	return n, err
}

// Close implements io.Closer, and closes the current logfile.
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.close()
}

// close closes the file if it is open.
func (l *Logger) close() error {
	if l.file == nil {
		return nil
	}
	err := l.file.Close()
	l.file = nil
	return err
}
