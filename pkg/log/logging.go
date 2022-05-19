package log

import (
	"log"

	"github.com/ory/x/errorsx"
)

var (
	currentLogLevel int
)

type Config struct {
	Level int `json:"level"`
}

// SetLogLevel :
func SetConfig(cfg *Config) {
	if cfg == nil {
		SetLogLevel(1)
		return
	}
	SetLogLevel(cfg.Level)
}

// SetLogLevel :
func SetLogLevel(level int) {
	currentLogLevel = level
}

// Logf :
func Logf(level int, a string, args ...interface{}) {
	if level <= currentLogLevel {
		log.Printf(a, args...)
	}

}

// Log :
func Log(level int, args ...interface{}) {
	if level <= currentLogLevel {
		log.Print(args...)
	}
}

// Logln :
func Logln(level int, args ...interface{}) {
	if level <= currentLogLevel {
		log.Println(args...)
	}
}

// Logln :
func Error(err error) bool {
	if err == nil {
		return false
	}

	err = errorsx.WithStack(err)

	if currentLogLevel == 0 {

		// include the stack trace in the error message
		Logf(0, "error: %+v", err)
	} else {
		Logf(1, "error: %v", err)
	}

	return true
}
