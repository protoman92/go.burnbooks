package goburnbooks

import "fmt"

// Logger logs some output.
type Logger interface {
	Printf(format string, args ...interface{})
}

type logger struct {
	enabled bool
}

func (l *logger) Printf(format string, args ...interface{}) {
	if l.enabled {
		newFormat := l.appendNewLine(format)
		fmt.Printf(newFormat, args...)
	}
}

func (l *logger) appendNewLine(format string) string {
	if len(format) > 0 {
		var newFormat string

		if format[len(format)-1] != '\n' {
			newFormat = format + "\n"
		} else {
			newFormat = format
		}

		return newFormat
	}

	return format
}

// NewLogger returns a new Logger.
func NewLogger(enabled bool) Logger {
	return &logger{enabled: enabled}
}
