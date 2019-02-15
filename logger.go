package subee

// Logger is the interface for logging
type Logger interface {
	Printf(format string, v ...interface{})
	Print(v ...interface{})
}
