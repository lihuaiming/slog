package main

import (
	"fmt"
	"github.com/lihuaiming/comm"
	"github.com/lihuaiming/slog"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"
)

func main() {
	c := make(chan os.Signal)
	signal.Notify(c)

	fmt.Println(comm.GetCurrPath() + "/log")
	slog.DefaultSLog(&slog.SlogCfg{
		LogDir:              comm.GetCurrPath() + "/log",
		Level:               "debug",
		LogFilePrefix:       "SynXXX",
		LogMaxFileSize:      1 * 1024 * 1024,
		LogMaxFileCountSize: 5 * 1024 * 1024,
	})

	if err := slog.DefaultLogStart(); err != nil {
		fmt.Println(err)
		return
	}
	go func() {
		time.Sleep(5 * time.Second)
		slog.DefaultSetLevel("Debug")
	}()

	defer slog.DefaultLogStop()

	slog.DefaultSetLevel("Debug")
	slog.Debug("lg.SetLevel == Debug")
	slog.Debug("Debug xxxxx 1")
	slog.Info("Info xxxxx 1")
	slog.Warn("Warn xxxxx 1")
	slog.Error("Error xxxxx 1")

	slog.DefaultSetLevel("Info")
	slog.Info("lg.SetLevel == Info")
	slog.Debug("Debug xxxxx 2")
	slog.Info("Info xxxxx 2")
	slog.Warn("Warn xxxxx 2")
	slog.Error("Error xxxxx 2")

	slog.DefaultSetLevel("Warn")
	slog.Warn("lg.SetLevel == Warn")
	slog.Debug("Debug xxxxx 3")
	slog.Info("Info xxxxx 3")
	slog.Warn("Warn xxxxx 3")
	slog.Error("Error xxxxx 3")

	slog.DefaultSetLevel("Error")
	slog.Error("lg.SetLevel == Error")
	slog.Debug("Debug xxxxx 3")
	slog.Info("Info xxxxx 3")
	slog.Warn("Warn xxxxx 3")
	slog.Error("Error xxxxx 3")

	slog.DefaultSetLevel("Debug")
	slog.Debug("------------------------")
	slog.Debug("Strat to write big logs ")
	slog.Debug("lg.SetLevel == Debug")
	slog.Debug("------------------------")

	slog.Debug("start")

	go func() {
		wg := &sync.WaitGroup{}
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(id int, w *sync.WaitGroup) {
				defer w.Done()
				for line := 1; line <= 10000; line++ {
					slog.Debug("  " + strconv.Itoa(line) + "  THREAD:" + strconv.Itoa(id) + "  " + comm.GetGuid())
					time.Sleep(10 * time.Millisecond)
				}
			}(i, wg)
		}
		wg.Wait()
	}()

	//signal check
	slog.Info("@main End signal: %s", (<-c).String())
}
