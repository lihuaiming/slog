package slog

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/sirupsen/logrus"
)

type SlogCfg struct {
	OutputSTD           bool
	Level               string
	LogFilePrefix       string
	LogDir              string
	LogKeepDays         int
	LogMaxFileSize      int64
	LogMaxFileCountSize int64

	MaxQueueSize int
}

type Slog struct {
	cfg    *SlogCfg //日志配置
	log    *logrus.Logger
	h      *SFileHook
	fmtter *STextFormatter

	caller_index int
}

func NewSlog(cfg *SlogCfg) *Slog {
	if cfg == nil {
		cfg = &SlogCfg{
			OutputSTD:           true,
			LogFilePrefix:       "slog",
			Level:               "debug",
			LogDir:              "log",
			LogKeepDays:         15,
			LogMaxFileSize:      100 * 1024 * 1024,       //100MB
			LogMaxFileCountSize: 10 * 1024 * 1024 * 1024, //10GB
			MaxQueueSize:        8192,
		}
	}

	//check
	if cfg.LogDir == "" {
		cfg.LogDir = "log"
	}
	le, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		le = logrus.TraceLevel
	}
	if cfg.LogKeepDays <= 0 {
		cfg.LogKeepDays = 15
	}
	if cfg.LogMaxFileSize <= 0 {
		cfg.LogMaxFileSize = 100 * 1024 * 1024 //100MB
	}
	if cfg.LogMaxFileCountSize <= 0 {
		cfg.LogMaxFileCountSize = 10 * 1024 * 1024 * 1024 //10GB
	} else {
		if cfg.LogMaxFileCountSize <= cfg.LogMaxFileSize {
			cfg.LogMaxFileCountSize = 2 * cfg.LogMaxFileSize
		}
	}
	if cfg.MaxQueueSize <= 0 {
		cfg.MaxQueueSize = 8192
	}

	newlog := &Slog{
		cfg: cfg,
		log: logrus.New(),
		h: NewSFileHook(&SFileHookCfg{
			Level:            le,
			FilePrefix:       cfg.LogFilePrefix,
			FileDir:          cfg.LogDir,
			KeepDays:         cfg.LogKeepDays,
			MaxFileSize:      cfg.LogMaxFileSize,
			MaxFileCountSize: cfg.LogMaxFileCountSize,
			MaxQueueSize:     cfg.MaxQueueSize,
		}),
		fmtter:       &STextFormatter{},
		caller_index: 2,
	}

	newlog.log.SetFormatter(newlog.fmtter)
	if cfg.OutputSTD {
		newlog.log.SetOutput(os.Stdout)
	}
	newlog.log.AddHook(newlog.h)
	newlog.log.SetLevel(le)

	return newlog
}

func (lg *Slog) SetLevel(level string) error {
	le, err := logrus.ParseLevel(level)
	if err != nil {
		return err
	}
	lg.log.SetLevel(le)
	return nil
}

func (lg *Slog) LogStart() error {
	err := lg.h.Init()
	if err == nil {
		lg.h.HookStart()
	}
	return err
}

func (lg *Slog) LogStop() {
	lg.h.HookStop()
}
func (lg *Slog) getline(i int) string {
	pc, file, line, ok := runtime.Caller(i)
	f := runtime.FuncForPC(pc)
	if !ok {
		return ""
	}
	return filepath.Base(file) + ":" + strconv.Itoa(line) + " " + f.Name() + "|"
}
func (lg *Slog) Debug(format string, v ...interface{}) {
	lg.log.Debug(fmt.Sprintf(lg.getline(lg.caller_index)+format, v...))
}
func (lg *Slog) Info(format string, v ...interface{}) {
	lg.log.Info(fmt.Sprintf(lg.getline(lg.caller_index)+format, v...))
}
func (lg *Slog) Warn(format string, v ...interface{}) {
	lg.log.Warn(fmt.Sprintf(lg.getline(lg.caller_index)+format, v...))
}
func (lg *Slog) Error(format string, v ...interface{}) {
	lg.log.Error(fmt.Sprintf(lg.getline(lg.caller_index)+format, v...))
}

/* 暂不开放
func (lg *Slog) Panic(format string, v ...interface{}) {
	lg.log.Panic(fmt.Sprintf(lg.getline(lg.caller_index)+format, v...))
}
*/
var Lg *Slog

func DefaultSLog(cfg *SlogCfg) error {
	if Lg == nil {
		Lg = NewSlog(cfg)
		Lg.caller_index += 1
	}
	return errors.New("DefaultSLog have another one !")
}

func DefaultSetLevel(level string) error {
	if Lg != nil {
		return Lg.SetLevel(level)
	} else {
		panic("DefaultLog is nil")
	}
}

func DefaultLogStart() error {
	if Lg != nil {
		return Lg.LogStart()
	} else {
		panic("DefaultLog is nil")
	}
}
func DefaultLogStop() {
	if Lg != nil {
		Lg.LogStop()
	} else {
		panic("DefaultLog is nil")
	}
}

func Debug(format string, v ...interface{}) {
	if Lg != nil {
		Lg.Debug(format, v...)
	} else {
		panic("DefaultLog is nil")
	}
}
func Info(format string, v ...interface{}) {
	if Lg != nil {
		Lg.Info(format, v...)
	} else {
		panic("DefaultLog is nil")
	}
}
func Warn(format string, v ...interface{}) {
	if Lg != nil {
		Lg.Warn(format, v...)
	} else {
		panic("DefaultLog is nil")
	}
}
func Error(format string, v ...interface{}) {
	if Lg != nil {
		Lg.Error(format, v...)
	} else {
		panic("DefaultLog is nil")
	}
}
func Panic() {
	if Lg != nil {
		if err := recover(); err != nil {
			var buf [4096]byte
			n := runtime.Stack(buf[:], false)
			Lg.Error("@Panic Panic ==>\n %s\n", string(buf[:n]))
			Lg.Error("@Panic %+v", err)
		}
	} else {
		panic("DefaultLog is nil")
	}
}

func AppExitWithWriteLog(lg *Slog) {
	if lg != nil {
		if err := recover(); err != nil {
			var buf [4096]byte
			n := runtime.Stack(buf[:], false)
			lg.Error("@AppExitWithWriteLog Panic ==>\n %s\n", string(buf[:n]))
			lg.Error("@AppExitWithWriteLog %+v", err)
		}
		lg.LogStop()
	}
	os.Exit(0)
}
func AppExitWithWriteLogEx(lg *Slog, code int) {
	if lg != nil {
		if err := recover(); err != nil {
			var buf [4096]byte
			n := runtime.Stack(buf[:], false)
			lg.Error("@AppExitWithWriteLog Panic ==>\n %s\n", string(buf[:n]))
			lg.Error("@AppExitWithWriteLog %+v", err)
		}
		lg.LogStop()
	}
	os.Exit(code)
}
