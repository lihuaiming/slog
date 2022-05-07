package slog

import (
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestListlogfile(t *testing.T) {
	hook := NewSFileHook(nil)
	hook.Init()

	//生成模拟数据
	e := time.Now()
	s := e.Add((-20) * 24 * time.Hour)
	for i := 0; i < 20; i++ {
		ioutil.WriteFile(hook.cfg.FileDir+"some"+"_"+s.Format("20060102")+"_1.log", []byte("test"), 0660)
		ioutil.WriteFile(hook.cfg.FileDir+hook.cfg.FilePrefix+"_"+s.Format("20060102")+"_1.log", []byte("test"), 0660)
		s = s.Add(24 * time.Hour)
	}

	sfiles := hook.listlogfile(hook.cfg.FileDir, hook.cfg.FilePrefix, ".log")
	if len(sfiles) == 0 {
		t.Failed()
	}
	t.Log(len(sfiles))
	for k, v := range sfiles {
		t.Log(k, v.path, v.size)
	}
	os.RemoveAll(hook.cfg.FileDir)

}

func TestGetFilePath(t *testing.T) {
	hook := NewSFileHook(&SFileHookCfg{
		KeepDays: 5,
	})
	hook.Init()
	//生成模拟数据
	e := time.Now()
	s := e.Add((-20) * 24 * time.Hour)
	for i := 0; i < 20; i++ {
		ioutil.WriteFile(hook.cfg.FileDir+hook.cfg.FilePrefix+"_"+s.Format("20060102")+"_1.log", []byte("test"), 0660)
		s = s.Add(24 * time.Hour)
	}

	logfile := hook.GetFilePath(time.Now(), 10)
	t.Log(logfile)
	logfile = hook.GetFilePath(time.Now().Add(24*time.Hour), 15)
	t.Log(logfile)

	sfiles := hook.listlogfile(hook.cfg.FileDir, hook.cfg.FilePrefix, ".log")
	if len(sfiles) == 0 {
		t.Failed()
	}
	t.Log(len(sfiles))
	for k, v := range sfiles {
		t.Log(k, v.path, v.size)
	}
	os.RemoveAll(hook.cfg.FileDir)
}

func TestCheckClear(t *testing.T) {
	hook := NewSFileHook(&SFileHookCfg{
		KeepDays: 5,
	})
	hook.Init()
	//生成模拟数据
	e := time.Now()
	s := e.Add((-20) * 24 * time.Hour)
	for i := 0; i < 20; i++ {
		s = s.Add(24 * time.Hour)
		ioutil.WriteFile(hook.cfg.FileDir+hook.cfg.FilePrefix+"_"+s.Format("20060102")+"_1.log", []byte("test"), 0660)
	}

	delfiles := hook.CheckClear(time.Now())
	for k, v := range delfiles {
		t.Log(k, v)
	}

	sfiles := hook.listlogfile(hook.cfg.FileDir, hook.cfg.FilePrefix, ".log")
	if len(sfiles) == 0 {
		t.Failed()
	}
	t.Log(len(sfiles))
	for k, v := range sfiles {
		t.Log(k, v.path, v.size)
	}
	os.RemoveAll(hook.cfg.FileDir)
}
