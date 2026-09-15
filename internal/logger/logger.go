package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Logger struct {
	dir string
	mu  sync.Mutex
}

func New() (*Logger, error) {
	dir, err := logDir()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	return &Logger{dir: dir}, nil
}

func logDir() (string, error) {
	if d := os.Getenv("LOCALAPPDATA"); d != "" {
		return filepath.Join(d, "HarborDownloader", "logs"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	switch {
	case fileExists(filepath.Join(home, "Library", "Logs")):
		return filepath.Join(home, "Library", "Logs", "HarborDownloader"), nil
	default:
		cache, err := os.UserCacheDir()
		if err != nil {
			return filepath.Join(home, ".harbor-downloader", "logs"), nil
		}
		return filepath.Join(cache, "HarborDownloader", "logs"), nil
	}
}

func fileExists(p string) bool {
	st, err := os.Stat(p)
	return err == nil && st.IsDir()
}

func (l *Logger) Info(msg string) { l.write("INFO", msg) }
func (l *Logger) Warn(msg string) { l.write("WARN", msg) }
func (l *Logger) Error(msg string) { l.write("ERROR", msg) }

func (l *Logger) write(level, msg string) {
	if l == nil {
		return
	}
	if looksSensitive(msg) {
		msg = "[redacted]"
	}
	line := fmt.Sprintf("%s %s %s\n", time.Now().Format("2006-01-02 15:04:05"), level, msg)
	l.mu.Lock()
	defer l.mu.Unlock()
	name := filepath.Join(l.dir, time.Now().Format("2006-01-02")+".log")
	f, err := os.OpenFile(name, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = f.WriteString(line)
}

func looksSensitive(msg string) bool {
	lower := strings.ToLower(msg)
	for _, k := range []string{"password", "token", "secret", "authorization:"} {
		if strings.Contains(lower, k) {
			return true
		}
	}
	return false
}

func (l *Logger) Dir() string {
	if l == nil {
		return ""
	}
	return l.dir
}
