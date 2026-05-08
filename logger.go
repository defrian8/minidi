package minidi

import "log"

// Simple logger
type Logger interface {
	Printf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

type defaultLogger struct{}

func (l *defaultLogger) Printf(format string, args ...interface{}) {
	log.Printf("[minidi] "+format, args...)
}

func (l *defaultLogger) Errorf(format string, args ...interface{}) {
	log.Printf("[minidi] ERROR: "+format, args...)
}
