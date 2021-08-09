package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"git.hub.com/wangyl/RTSP_AGREEMENT/Global"
	"git.hub.com/wangyl/RTSP_AGREEMENT/internal/RTSP"
	"git.hub.com/wangyl/RTSP_AGREEMENT/pkg/Settings"
	"gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	a := kingpin.New(filepath.Base(os.Args[0]), "rtsp server")
	a.HelpFlag.Short('h')
	a.Flag("config", "config path").Short('c').StringVar(&Global.ConfigPath)
	if _, err := a.Parse(os.Args[1:]); err != nil {
		fmt.Println("Parse Cmd Param fail:", err)
		os.Exit(-1)
	}
	if err := Global.GlobalInit(); err != nil {
		fmt.Println("Global Init fail:", err)
		os.Exit(-1)
	}
	//start service
	var srv = RTSP.NewRtspServer(RTSP.Option{Cfg: Settings.GetConfig().RtspServer})
	if err := srv.Serve(); err != nil {
		fmt.Println("Start Rtsp Server Fail:", err)
		os.Exit(-1)
	}
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	s := <-quit
	switch s {
	case syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT:
		srv.Stop()
	}
}
