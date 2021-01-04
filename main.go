package main

import (
	"git.hub.com/wangyl/RTSP_AGREEMENT/app"
	"git.hub.com/wangyl/RTSP_AGREEMENT/pkg/Logger"
	"git.hub.com/wangyl/RTSP_AGREEMENT/pkg/Settings"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

var configPath string

func main() {
	a := kingpin.New(filepath.Base(os.Args[0]), "rtsp service")
	a.HelpFlag.Short('h')
	a.Flag("config", "config path").Short('c').StringVar(&configPath)
	if _, err := a.Parse(os.Args[1:]); err != nil {
		Logger.GetLogger().Error("init flag fail: " + err.Error())
		os.Exit(-1)
	}
	//init config
	if err := Settings.ReadConfig(configPath); err != nil {
		Logger.GetLogger().Error("init config fail: " + err.Error())
		os.Exit(-1)
	}
	//start service
	var rtspService app.RtspService
	rtspService.Init(Settings.GetConfig().App.RtspPort)
	rtspService.StartWork()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	s := <-quit
	switch s {
	case syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT:
		rtspService.Stop()
	}
}
