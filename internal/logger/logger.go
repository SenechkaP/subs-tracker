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

var Log = logrus.New()

func init() {
	Log.SetFormatter(&PlainFormatter{})
	Log.SetLevel(logrus.InfoLevel)
}
