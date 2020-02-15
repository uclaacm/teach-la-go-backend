package log

import (
	"fmt"
	"os"
	"time"
)

// LogLevel is the type of a logging level.
type LogLevel uint

// Logging levels are enumerated in descending order of importance
const (
	// Fatal describes a logging level at which
	// anything logged will cause a program-halting
	// error.
	// Using Fatal will halt execution of the program.
	Fatal = iota

	// Critical describes a logging level at which
	// there is a high likelihood the event logged
	// will have program-halting repercussions.
	Critical

	// Error describes a logging level at which
	// the program has encountered an unexpected
	// situation.
	Error

	// Info describes a logging level at which the
	// program simply wishes to notify the user
	// of basic events.
	Info

	// Debug describes a verbose logging level
	// in which all relevant information is
	// logged for a program.
	Debug

	// Levels is the number of available verbosity
	// levels.
	Levels
)

// LogLevelPrefix acquires the prefix string for
// all logging messages of level l.
func logLevelPrefix(l LogLevel) string {
	switch l {
	case Fatal:
		return "FATAL: "
	case Critical:
		return "CRITICAL: "
	case Error:
		return "ERROR: "
	case Info:
		return "INFO: "
	case Debug:
		return "DEBUG: "
	default:
		return ""
	}
}

var Verbosity LogLevel

func init() {
	Verbosity = Debug
}

func logWithLevel(level LogLevel, format string, v ...interface{}) {
	// ignore messages that are more verbose than the current
	// level. This eliminates the need for enum bounds checking.
	if level > Verbosity {
		return
	}

	prefix := logLevelPrefix(level)
	time := time.Now().Format("2006/01/02 15:04:05")

	message := fmt.Sprintf("%s %s"+format+"\n", append([]interface{}{time, prefix}, v...)...)

	// Avoids EXTRA printf error
	if len(v) == 0 {
		message = fmt.Sprintf("%s %s"+format+"\n", time, prefix)
	}

	fmt.Println(message)

	if level == Fatal {
		os.Exit(1)
	}
}

func Println(v ...interface{}) {
	fmt.Println(v...)
}

// Debugf outputs the format with level Debug to stdout
func Debugf(format string, v ...interface{}) {
	logWithLevel(Debug, format, v...)
}

// Infof outputs the format with level Info to stdout
func Infof(format string, v ...interface{}) {
	logWithLevel(Info, format, v...)
}

// Errorf outputs the format with level Error to stdout
func Errorf(format string, v ...interface{}) {
	logWithLevel(Error, format, v...)
}

// Criticalf outputs the format with level Critical to stdout
func Criticalf(format string, v ...interface{}) {
	logWithLevel(Critical, format, v...)
}

func Fatalf(format string, v ...interface{}) {
	logWithLevel(Fatal, format, v...)
}

// setVerbosity changes the minimum verbosity to print
func SetVerbosity(level LogLevel) {
	Verbosity = level
}
