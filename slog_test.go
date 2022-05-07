package slog

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestLogrusSTextFormatter_Format(t *testing.T) {
	logrus.SetFormatter(&STextFormatter{})
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.TraceLevel)
	logrus.WithFields(logrus.Fields{
		"animal": "walrus",
		"size":   10,
	}).Info("A group of walrus emerges from the ocean")
	logrus.WithFields(logrus.Fields{
		"omg":    true,
		"number": 122,
	}).Warn("The group's number increased tremendously!")
	logrus.Warn("Maxwell")

}

func TestSlog(t *testing.T) {
	lg := NewSlog(&SlogCfg{
		Level:          "debug",
		LogFilePrefix:  "SynXXX",
		LogMaxFileSize: 1 * 1024 * 1024,
	})
	t.Logf("%+v", lg.h.cfg)

	if err := lg.LogStart(); err != nil {
		t.Fatal(err)
	}

	defer lg.LogStop()

	lg.Debug("Debug xxxxx")
	lg.Info("Info xxxxx")
	lg.Warn("Warn xxxxx")
	lg.Error("Error xxxxx")
	//lg.Panic("Panic xxxxx")

}
