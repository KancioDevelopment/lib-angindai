package logging

import (
	"os"

	"github.com/sirupsen/logrus"
)

type (
	Logger struct {
		Out *logrus.Logger
		Err *logrus.Logger
	}
)

func New() *Logger {
	return &Logger{
		Out: &logrus.Logger{
			Formatter: new(logrus.TextFormatter),
			Out:       os.Stdout,
			Level:     logrus.InfoLevel,
		},
		Err: &logrus.Logger{
			Formatter: new(logrus.TextFormatter),
			Out:       os.Stderr,
			Level:     logrus.InfoLevel,
		},
	}
}
