package logging

import "github.com/sirupsen/logrus"

func NewDefaultLogger() *logrus.Logger {
	logger := logrus.New()
	customFormatter := new(logrus.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	customFormatter.FullTimestamp = true
	logger.SetFormatter(customFormatter)
	return logger
}
