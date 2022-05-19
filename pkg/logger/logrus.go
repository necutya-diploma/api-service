package logger

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	jsonLogFormat = "json"
)

func Init(logFormat, logLevel string) {
	switch logFormat {
	case jsonLogFormat:
		logrus.SetFormatter(&logrus.JSONFormatter{})
	default:
		logrus.SetFormatter(&logrus.TextFormatter{})
	}

	logrus.SetOutput(os.Stderr)

	parsedLevel, err := logrus.ParseLevel(strings.ToLower(logLevel))
	if err != nil {
		parsedLevel = logrus.DebugLevel
	}
	logrus.SetLevel(parsedLevel)
}

func Debug(msg ...interface{}) {
	logrus.Debug(msg...)
}

func Debugf(format string, args ...interface{}) {
	logrus.Debugf(format, args...)
}

func Info(msg ...interface{}) {
	logrus.Info(msg...)
}

func Infof(format string, args ...interface{}) {
	logrus.Infof(format, args...)
}

func Warn(msg ...interface{}) {
	logrus.Warn(msg...)
}

func Warnf(format string, args ...interface{}) {
	logrus.Warnf(format, args...)
}

func Error(msg ...interface{}) {
	logrus.Error(msg...)
}

func Errorf(format string, args ...interface{}) {
	logrus.Errorf(format, args...)
}

func Fatal(msg ...interface{}) {
	logrus.Fatal(msg...)
}

func Fatalf(format string, args ...interface{}) {
	logrus.Fatalf(format, args...)
}
