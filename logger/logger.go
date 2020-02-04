package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

// LogLevel is the type of a logging level.
type LogLevel uint

// Logging levels are enumerated in descending order
// such that the higher the code of the logging level,
// the higher the verbosity of the Slogger.
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
func LogLevelPrefix(l LogLevel) string {
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

// Slogger is a simple Logger.
type Slogger struct {
	*log.Logger
	verbosity   LogLevel
	usingStdout bool
	out         io.Writer
}

// New returns a new Slogger with the specified settings.
func New(out io.Writer, prefix string, flag int, verbosity LogLevel) *Slogger {
	s := &Slogger{}

	s.Logger = log.New(out, prefix, flag)

	s.out = out
	s.verbosity = verbosity

	return s
}

func (s *Slogger) logWithLevel(level LogLevel, format string, v ...interface{}) {
	// ignore messages that are more verbose than the current
	// level. This eliminates the need for enum bounds checking.
	if level > s.verbosity {
		return
	}

	prefix := LogLevelPrefix(level)
	time := time.Now().Format("2006/01/02 15:04:05")

	message := fmt.Sprintf("%s %s"+format+"\n", append([]interface{}{time, prefix}, v...)...)

	// Avoids EXTRA printf error
	if len(v) == 0 {
		message = fmt.Sprintf("%s %s"+format+"\n", time, prefix)
	}

	// Log the message to the Slogger's output device
	io.WriteString(s.out, message)
}

// Debugf outputs the format with level Debug to its Slogger's output
func (s *Slogger) Debugf(format string, v ...interface{}) {
	s.logWithLevel(Debug, format, v...)
}

// Infof outputs the format with level Info to its Slogger's output
func (s *Slogger) Infof(format string, v ...interface{}) {
	s.logWithLevel(Info, format, v...)
}

// Errorf outputs the format with level Error to its Slogger's output
func (s *Slogger) Errorf(format string, v ...interface{}) {
	s.logWithLevel(Error, format, v...)
}

// Criticalf outputs the format with level Critical to its Slogger's output
func (s *Slogger) Criticalf(format string, v ...interface{}) {
	s.logWithLevel(Critical, format, v...)
}

// setVerbosity changes the verbosity of the logger
func (s *Slogger) setVerbosity(level LogLevel) {
	s.verbosity = level
}

// setOutput changes the output of the logger
func (s *Slogger) setOutput(out io.Writer) {
	s.out = out
}

// Create a default Slogger
var logger = New(os.Stdout, "", log.LstdFlags, Debug)

// Printf logs the formatted string to the default slogger
// (stdout).
func Printf(format string, v ...interface{}) {
	time := time.Now().Format("2006/01/02 15:04:05")
	fmt.Printf(format, append([]interface{}{time}, v...)...)
}

// Println logs the formatted string to the default slogger
// (stdout).
func Println(v ...interface{}) {
	time := time.Now().Format("2006/01/02 15:04:05")
	fmt.Println(append([]interface{}{time}, v...)...)
}

// Calls Fatalf on the default logger
func Fatalf(format string, v ...interface{}) {
	logger.Fatalf(format, v...)
}

// Calls Fatalln on the default logger
func Fatalln(v ...interface{}) {
	logger.Fatalln(v...)
}

// Calls Debugf on the default logger
func Debugf(format string, v ...interface{}) {
	logger.Debugf(format, v...)
}

// Calls Infof on the default logger
func Infof(format string, v ...interface{}) {
	logger.Infof(format, v...)
}

// Calls Errorf on the default logger
func Errorf(format string, v ...interface{}) {
	logger.Errorf(format, v...)
}

// Calls Criticalf on the default logger
func Criticalf(format string, v ...interface{}) {
	logger.Criticalf(format, v...)
}

// Calls setVerbosity on the default logger
func setVerbosity(level LogLevel) {
	logger.setVerbosity(level)
}

// Calls setOutput on the default logger
func setOutput(out io.Writer) {
	logger.setOutput(out)
}
