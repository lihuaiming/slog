package slog

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
)

type SFileHookMsg struct {
	t    time.Time
	data []byte
}

type SFileHookCfg struct {
	Level            logrus.Level
	FilePrefix       string
	FileDir          string
	KeepDays         int
	MaxFileSize      int64
	MaxFileCountSize int64

	MaxQueueSize int //队列长度
}

type SFileHook struct {
	//base
	cfg *SFileHookCfg
	//
	msgs chan *SFileHookMsg

	Enable bool

	//Runtime
	lastFileDay  int
	lastFileCo   int
	lastFileName string
	lastFileSize int64

	//signal
	wg *sync.WaitGroup
}

func NewSFileHook(cfg *SFileHookCfg) *SFileHook {
	h := &SFileHook{
		lastFileDay:  0,
		lastFileCo:   0,
		lastFileName: "",
		lastFileSize: 0,
	}
	if cfg == nil {
		h.cfg = &SFileHookCfg{}
	} else {
		h.cfg = cfg
	}
	//check
	if h.cfg.FilePrefix == "" {
		h.cfg.FilePrefix = "slog"
	}
	if h.cfg.FileDir == "" {
		h.cfg.FileDir = "log"
	}
	if h.cfg.KeepDays <= 0 {
		h.cfg.KeepDays = 15
	}
	if h.cfg.MaxFileSize <= 0 {
		h.cfg.MaxFileSize = 100 * 1024 * 1024 //100MB
	}
	if h.cfg.MaxFileCountSize <= 0 {
		h.cfg.MaxFileCountSize = 10 * 1024 * 1024 * 1024 //10GB
	} else {
		if h.cfg.MaxFileCountSize <= h.cfg.MaxFileSize {
			h.cfg.MaxFileCountSize = 2 * h.cfg.MaxFileSize
		}
	}
	if h.cfg.MaxQueueSize <= 0 {
		h.cfg.MaxQueueSize = 8192
	}

	h.msgs = make(chan *SFileHookMsg, h.cfg.MaxQueueSize)
	h.wg = &sync.WaitGroup{}
	return h
}

func (hook *SFileHook) Fire(entry *logrus.Entry) error {
	line, err := entry.String()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to read entry, %v", err)
		return err
	}
	if len(hook.msgs) >= hook.cfg.MaxQueueSize {
		fmt.Println("too many logs.")
		return errors.New("too many logs.")
	}
	//fmt.Println("SFileHook.Fire " + line)
	hook.msgs <- &SFileHookMsg{
		t:    entry.Time,
		data: []byte(line),
	}
	return nil
}

func (hook *SFileHook) Levels() []logrus.Level {
	return logrus.AllLevels

}

func (hook *SFileHook) Init() error {
	//确保路径正确 且存在
	hook.cfg.FileDir = filepath.Clean(hook.cfg.FileDir)
	if hook.cfg.FileDir[len(hook.cfg.FileDir)-1:] == "/" {
		hook.cfg.FileDir = hook.cfg.FileDir[:len(hook.cfg.FileDir)-1]
	}
	os.MkdirAll(hook.cfg.FileDir, os.ModePerm)
	hook.cfg.FileDir += "/"
	hook.Enable = true
	return nil
}

type searchlogfile struct {
	path string
	size int64
}

func (hook *SFileHook) listlogfile(dir string, prefix string, ext string) []*searchlogfile {
	sfs := make([]*searchlogfile, 0, 1024)
	filepath.Walk(dir, func(filename string, info os.FileInfo, err error) error {
		if info.IsDir() {

		} else {
			if filepath.Ext(filepath.Base(filename)) == ext {
				filenameWithSuffix := filepath.Base(filename)
				fileSuffix := filepath.Ext(filenameWithSuffix)
				fileOnlyName := strings.TrimSuffix(filenameWithSuffix, fileSuffix)
				items := strings.Split(fileOnlyName, "_")
				if len(items) >= 3 {
					if items[0] == prefix {
						sfs = append(sfs, &searchlogfile{
							path: filename,
							size: info.Size(),
						})
					}

				}

			}
		}
		return nil
	})

	return sfs
}

func (hook *SFileHook) GetFilePath(t time.Time, needbytes int64) string {
	//初始化
	if hook.lastFileDay == 0 {
		//make sure which day
		hook.lastFileDay = t.Day()
		//list log files
		sfiles := hook.listlogfile(hook.cfg.FileDir, hook.cfg.FilePrefix, ".log")
		//tmp
		lastFileCo := -1
		lastFileSize := int64(0)
		for index := 0; index < len(sfiles); index++ {
			filenameWithSuffix := filepath.Base(sfiles[index].path)
			fileSuffix := filepath.Ext(filenameWithSuffix)
			fileOnlyName := strings.TrimSuffix(filenameWithSuffix, fileSuffix)
			items := strings.Split(fileOnlyName, "_")
			if len(items) >= 3 {
				if items[1] == t.Format("20060102") {
					co, err := strconv.Atoi(items[2])
					if err == nil && co > lastFileCo {
						lastFileCo = co
						lastFileSize = sfiles[index].size
					}
				}
			}
		}
		if lastFileCo == -1 {
			hook.lastFileCo = 0
			hook.lastFileName = hook.cfg.FileDir + hook.cfg.FilePrefix + "_" + t.Format("20060102") + "_" + strconv.Itoa(hook.lastFileCo) + ".log"
			hook.lastFileSize = 0
		} else {
			hook.lastFileCo = lastFileCo
			hook.lastFileName = hook.cfg.FileDir + hook.cfg.FilePrefix + "_" + t.Format("20060102") + "_" + strconv.Itoa(hook.lastFileCo) + ".log"
			hook.lastFileSize = lastFileSize
		}
	}
	nowLogFile := ""
	bCheckFilesClean := false
	//文件选择
	FileDay := t.Day()
	if FileDay != hook.lastFileDay { //change day
		bCheckFilesClean = true

		hook.lastFileDay = FileDay
		hook.lastFileCo = 0
		hook.lastFileName = hook.cfg.FileDir + hook.cfg.FilePrefix + "_" + t.Format("20060102") + "_" + strconv.Itoa(hook.lastFileCo) + ".log"
		hook.lastFileSize = 0
		nowLogFile = hook.lastFileName
	} else { //
		if needbytes+hook.lastFileSize > hook.cfg.MaxFileSize {
			hook.lastFileCo = hook.lastFileCo + 1
			hook.lastFileName = hook.cfg.FileDir + hook.cfg.FilePrefix + "_" + t.Format("20060102") + "_" + strconv.Itoa(hook.lastFileCo) + ".log"
			hook.lastFileSize = 0
			nowLogFile = hook.lastFileName
			bCheckFilesClean = true
		} else {
			nowLogFile = hook.lastFileName
		}
	}

	//数据清理check
	if bCheckFilesClean {
		hook.CheckClear(t)
	}

	return nowLogFile
}
func (hook *SFileHook) CheckClear(t time.Time) []string {
	delfiles := make([]string, 0, 64)
	//list log files
	sfiles := hook.listlogfile(hook.cfg.FileDir, hook.cfg.FilePrefix, ".log")
	sort.Slice(sfiles, func(i, j int) bool {
		filenameWithSuffix1 := filepath.Base(sfiles[i].path)
		fileSuffix1 := filepath.Ext(filenameWithSuffix1)
		fileOnlyName1 := strings.TrimSuffix(filenameWithSuffix1, fileSuffix1)

		filenameWithSuffix2 := filepath.Base(sfiles[j].path)
		fileSuffix2 := filepath.Ext(filenameWithSuffix2)
		fileOnlyName2 := strings.TrimSuffix(filenameWithSuffix2, fileSuffix2)

		items1 := strings.Split(fileOnlyName1, "_")
		items2 := strings.Split(fileOnlyName2, "_")

		if len(items1) >= 3 && len(items2) >= 3 {
			day1, _ := strconv.Atoi(items1[1])
			day2, _ := strconv.Atoi(items2[1])
			if day1 < day2 {
				return true
			} else if day1 == day2 {
				co1, _ := strconv.Atoi(items1[2])
				co2, _ := strconv.Atoi(items2[2])
				return co1 < co2

			} else {
				return false
			}
		}

		return sfiles[i].path < sfiles[j].path
	})

	totleSize := int64(0)
	//时间
	expireday := t.Add(time.Duration(hook.cfg.KeepDays) * (-24) * time.Hour)
	for i := 0; i < len(sfiles); i++ {
		filenameWithSuffix := filepath.Base(sfiles[i].path)
		fileSuffix := filepath.Ext(filenameWithSuffix)
		fileOnlyName := strings.TrimSuffix(filenameWithSuffix, fileSuffix)
		items := strings.Split(fileOnlyName, "_")
		if len(items) >= 3 {
			if items[1] <= expireday.Format("20060102") {
				os.Remove(sfiles[i].path)
				delfiles = append(delfiles, sfiles[i].path)
				sfiles[i].size = 0
			}
		}
		totleSize += sfiles[i].size
	}
	//空间
	for i := 0; i < len(sfiles); i++ {
		if totleSize < hook.cfg.MaxFileCountSize {
			break
		}
		if sfiles[i].size > 0 {
			os.Remove(sfiles[i].path)
			delfiles = append(delfiles, sfiles[i].path)
		}
		totleSize = totleSize - sfiles[i].size
	}
	return delfiles
}

// write to os.File.
func (hook *SFileHook) Write(t time.Time, b []byte) (int, error) {
	fd, err := os.OpenFile(hook.GetFilePath(t, int64(len(b))), os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0660)
	if err == nil {
		defer fd.Close()
		n, err := fd.Write(b)
		atomic.AddInt64(&hook.lastFileSize, int64(n))
		return n, err
	}
	return 0, err
}

func (hook *SFileHook) HookStart() {
	go hook.GoWriter()
}

//GoWriter 间隔写入携程
func (hook *SFileHook) GoWriter() {
	/*
		fmt.Println("SFileHook::GoWriter start")
		defer func() {
			fmt.Println("SFileHook::GoWriter end")
		}()
	*/
	hook.wg.Add(1)
	defer hook.wg.Done()
	for {
		msg, ok := <-hook.msgs
		//fmt.Println("SFileHook::GoWriter " + string(msg.data))
		if ok {
			//fmt.Println("SFileHook::GoWriter write")
			hook.Write(msg.t, msg.data)
		} else {
			//fmt.Println("SFileHook::GoWriter break")
			break
		}

	}
}

func (hook *SFileHook) HookStop() {
	hook.Enable = false
	close(hook.msgs)
	hook.wg.Wait()
}
