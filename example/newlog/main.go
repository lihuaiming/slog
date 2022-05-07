package main

import (
	"fmt"
	"github.com/lihuaiming/comm"
	"github.com/lihuaiming/slog"
	"os"
	"os/signal"
	"time"
)

func main() {
	c := make(chan os.Signal)
	signal.Notify(c)

	fmt.Println(comm.GetCurrPath() + "/log")
	lg := slog.NewSlog(&slog.SlogCfg{
		LogDir:         comm.GetCurrPath() + "/log",
		Level:          "debug",
		LogFilePrefix:  "SynXXX",
		LogMaxFileSize: 1 * 1024 * 1024,
	})

	if err := lg.LogStart(); err != nil {
		fmt.Println(err)
		return
	}

	defer lg.LogStop()

	lg.SetLevel("Debug")
	lg.Debug("lg.SetLevel == Debug")
	lg.Debug("Debug xxxxx 1")
	lg.Info("Info xxxxx 1")
	lg.Warn("Warn xxxxx 1")
	lg.Error("Error xxxxx 1")

	lg.SetLevel("Info")
	lg.Info("lg.SetLevel == Info")
	lg.Debug("Debug xxxxx 2")
	lg.Info("Info xxxxx 2")
	lg.Warn("Warn xxxxx 2")
	lg.Error("Error xxxxx 2")

	lg.SetLevel("Warn")
	lg.Warn("lg.SetLevel == Warn")
	lg.Debug("Debug xxxxx 3")
	lg.Info("Info xxxxx 3")
	lg.Warn("Warn xxxxx 3")
	lg.Error("Error xxxxx 3")

	lg.SetLevel("Error")
	lg.Error("lg.SetLevel == Error")
	lg.Debug("Debug xxxxx 3")
	lg.Info("Info xxxxx 3")
	lg.Warn("Warn xxxxx 3")
	lg.Error("Error xxxxx 3")

	time.Sleep(15 * time.Second)
	//signal check
	slog.Info("@main End signal: %s", (<-c).String())

}
