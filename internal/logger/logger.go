package logger

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

type PlainFormatter struct{}

func (f *PlainFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestamp := entry.Time.Format("2006/01/02 15:04:05")

	level := entry.Level.String()

	log := fmt.Sprintf("%s %s %s\n",
		timestamp,
		level,
		entry.Message,
	)

	return []byte(log), nil
}

type AccessFormatter struct{}

func (f *AccessFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestamp := entry.Time.Format("2006/01/02 15:04:05")
	return fmt.Appendf(nil, "%s %s\n", timestamp, entry.Message), nil
}

var Log = logrus.New()
var Access = logrus.New()

func init() {
	Log.SetFormatter(&PlainFormatter{})
	Log.SetLevel(logrus.InfoLevel)

	Access.SetFormatter(&AccessFormatter{})
}
