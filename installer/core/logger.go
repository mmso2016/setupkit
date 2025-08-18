package core

// Logger interface for logging operations
type Logger interface {
	Debug(msg string, keysAndValues ...interface{})
	Info(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
	Verbose(msg string, keysAndValues ...interface{})
	VerboseSection(section string)
	SetVerbose(verbose bool)
	Close() error
}
