package slog

import (
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestSTextFormatter_Format(t *testing.T) {
	fmt := &STextFormatter{}
	en := &logrus.Entry{
		Data:    logrus.Fields{},
		Time:    time.Now(),
		Level:   logrus.DebugLevel,
		Message: "hello world!",
	}
	buf, err := fmt.Format(en)
	if err != nil {
		t.Error(err)
	} else {
		t.Log(string(buf))
	}

}
