package Global

import (
	"git.hub.com/wangyl/MediaSreamServer/pkg/Logger"
	"git.hub.com/wangyl/MediaSreamServer/pkg/Settings"
)

var ConfigPath string

func GlobalInit() (err error) {
	//init config
	if err = Settings.ReadConfig(ConfigPath); err != nil {
		return err
	}
	//init Logger
	if err = Logger.Init(
		Logger.SetDevelopment(Settings.GetConfig().Logger.Development),
		Logger.SetMaxAge(Settings.GetConfig().Logger.MaxAge),
		Logger.SetMaxBackups(Settings.GetConfig().Logger.MaxBackups),
		Logger.SetMaxSize(Settings.GetConfig().Logger.MaxSize),
	); err != nil {
		return err
	}
	return
}
