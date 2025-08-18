package core

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
)

// SimpleLogger is a basic logger implementation
type SimpleLogger struct {
	level    LogLevel
	verbose  bool
	logger   *log.Logger
	file     *os.File
	mu       sync.Mutex
}

// LogLevel represents logging levels
type LogLevel int

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	VerboseLevel
)

// NewLogger creates a new logger instance
func NewLogger(level string, logFile string) Logger {
	var logLevel LogLevel
	switch strings.ToLower(level) {
	case "debug":
		logLevel = DebugLevel
	case "info":
		logLevel = InfoLevel
	case "warn", "warning":
		logLevel = WarnLevel
	case "error":
		logLevel = ErrorLevel
	default:
		logLevel = InfoLevel
	}
	
	l := &SimpleLogger{
		level:   logLevel,
		verbose: false,
	}
	
	var output io.Writer = os.Stdout
	
	if logFile != "" {
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err == nil {
			l.file = file
			// Write to both file and stdout
			output = io.MultiWriter(os.Stdout, file)
		}
	}
	
	l.logger = log.New(output, "", log.LstdFlags)
	
	return l
}

// Debug logs a debug message
func (l *SimpleLogger) Debug(msg string, keysAndValues ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	if l.level <= DebugLevel {
		l.log("DEBUG", msg, keysAndValues...)
	}
}

// Info logs an info message
func (l *SimpleLogger) Info(msg string, keysAndValues ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	if l.level <= InfoLevel {
		l.log("INFO", msg, keysAndValues...)
	}
}

// Warn logs a warning message
func (l *SimpleLogger) Warn(msg string, keysAndValues ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	if l.level <= WarnLevel {
		l.log("WARN", msg, keysAndValues...)
	}
}

// Error logs an error message
func (l *SimpleLogger) Error(msg string, keysAndValues ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	if l.level <= ErrorLevel {
		l.log("ERROR", msg, keysAndValues...)
	}
}

// Verbose logs a verbose message
func (l *SimpleLogger) Verbose(msg string, keysAndValues ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	if l.verbose {
		l.log("VERBOSE", msg, keysAndValues...)
	}
}

// VerboseSection logs a verbose section header
func (l *SimpleLogger) VerboseSection(section string) {
	if l.verbose {
		l.mu.Lock()
		defer l.mu.Unlock()
		l.logger.Printf("  %s", section)
	}
}

// SetVerbose enables or disables verbose logging
func (l *SimpleLogger) SetVerbose(verbose bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.verbose = verbose
}

// Close closes the log file if open
func (l *SimpleLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	if l.file != nil {
		err := l.file.Close()
		l.file = nil
		return err
	}
	return nil
}

// log formats and writes a log message
func (l *SimpleLogger) log(level string, msg string, keysAndValues ...interface{}) {
	// Format key-value pairs
	var extras string
	if len(keysAndValues) > 0 {
		var pairs []string
		for i := 0; i < len(keysAndValues)-1; i += 2 {
			key := fmt.Sprintf("%v", keysAndValues[i])
			value := fmt.Sprintf("%v", keysAndValues[i+1])
			pairs = append(pairs, fmt.Sprintf("%s=%s", key, value))
		}
		if len(pairs) > 0 {
			extras = " " + strings.Join(pairs, " ")
		}
	}
	
	l.logger.Printf("[%s] %s%s", level, msg, extras)
}

// NullLogger is a logger that discards all output
type NullLogger struct{}

// NewNullLogger creates a logger that discards all output
func NewNullLogger() Logger {
	return &NullLogger{}
}

func (n *NullLogger) Debug(msg string, keysAndValues ...interface{})        {}
func (n *NullLogger) Info(msg string, keysAndValues ...interface{})         {}
func (n *NullLogger) Warn(msg string, keysAndValues ...interface{})         {}
func (n *NullLogger) Error(msg string, keysAndValues ...interface{})        {}
func (n *NullLogger) Verbose(msg string, keysAndValues ...interface{})      {}
func (n *NullLogger) VerboseSection(section string)                         {}
func (n *NullLogger) SetVerbose(verbose bool)                              {}
func (n *NullLogger) Close() error                                         { return nil }
