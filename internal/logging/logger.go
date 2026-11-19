package logging

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Logger struct {
	logger   *slog.Logger
	logFile  *os.File
	logLines []string
	logMu    sync.RWMutex
	maxLines int
}

func New(logPath string, maxLogLines int) (*Logger, error) {
	// Ensure log directory exists
	logDir := filepath.Dir(logPath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open log file
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	// Create multi-writer for both stdout and file
	multiWriter := io.MultiWriter(os.Stdout, logFile)

	logger := &Logger{
		logger:   slog.New(slog.NewJSONHandler(multiWriter, nil)),
		logFile:  logFile,
		logLines: make([]string, 0, maxLogLines),
		maxLines: maxLogLines,
	}

	return logger, nil
}

func (l *Logger) Debug(msg string, args ...interface{}) {
	l.logger.Debug(msg, toAttrs(args)...)
	l.recordLog("DEBUG", msg)
}

func (l *Logger) Info(msg string, args ...interface{}) {
	l.logger.Info(msg, toAttrs(args)...)
	l.recordLog("INFO", msg)
}

func (l *Logger) Warn(msg string, args ...interface{}) {
	l.logger.Warn(msg, toAttrs(args)...)
	l.recordLog("WARN", msg)
}

func (l *Logger) Error(msg string, args ...interface{}) {
	l.logger.Error(msg, toAttrs(args)...)
	l.recordLog("ERROR", msg)
}

func (l *Logger) recordLog(level, msg string) {
	l.logMu.Lock()
	defer l.logMu.Unlock()

	logEntry := fmt.Sprintf("[%s] %s: %s", time.Now().Format(time.RFC3339), level, msg)
	l.logLines = append(l.logLines, logEntry)

	// Keep only recent logs
	if len(l.logLines) > l.maxLines {
		l.logLines = l.logLines[len(l.logLines)-l.maxLines:]
	}
}

func (l *Logger) GetRecentLogs(count int) []string {
	l.logMu.RLock()
	defer l.logMu.RUnlock()

	if count > len(l.logLines) {
		count = len(l.logLines)
	}

	result := make([]string, count)
	copy(result, l.logLines[len(l.logLines)-count:])
	return result
}

func (l *Logger) Close() error {
	if l.logFile != nil {
		return l.logFile.Close()
	}
	return nil
}

func toAttrs(args []interface{}) []slog.Attr {
	var attrs []slog.Attr
	for i := 0; i < len(args)-1; i += 2 {
		if key, ok := args[i].(string); ok {
			attrs = append(attrs, slog.Any(key, args[i+1]))
		}
	}
	return attrs
}
