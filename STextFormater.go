package slog

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

// TextFormatter formats logs into text
type STextFormatter struct {
}

// Format renders a single log entry
func (f *STextFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	str := ""
	if len(entry.Data) > 0 {
		str = fmt.Sprintf("%s|%d|%s|%s|%+v\n",
			entry.Time.Format("20060102|150405"),
			entry.Time.Nanosecond()/1000000,
			entry.Level,
			entry.Message,
			entry.Data,
		)
	} else {
		str = fmt.Sprintf("%s|%d|%s|%s\n",
			entry.Time.Format("20060102|150405"),
			entry.Time.Nanosecond()/1000000,
			entry.Level,
			entry.Message,
		)
	}

	return []byte(str), nil
}
