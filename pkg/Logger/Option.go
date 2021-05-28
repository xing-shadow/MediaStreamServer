package Logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path/filepath"
)

type Option struct {
	LogDir      string
	FiLeName    string
	Level       zapcore.Level
	MaxSize     int // 单位MB
	MaxBackups  int // 旧文件最大数量
	MaxAge      int // 旧文件最大保存时间 单位:天
	Development bool
	zap.Config
}

func (i *Option) fixup() {
	if i.LogDir == "" {
		path, _ := os.Executable()
		i.LogDir = filepath.Dir(path)
	}
	if i.FiLeName == "" {
		i.FiLeName = filepath.Base(os.Args[0])
	}
	if i.MaxBackups == 0 {
		i.MaxBackups = 5
	}
	if i.MaxSize == 0 {
		i.MaxSize = 500
	}
	if i.MaxAge == 0 {
		i.MaxAge = 5
	}
}

type ModOptions func(options *Option)

func SetMaxSize(MaxSize int) ModOptions {
	return func(option *Option) {
		option.MaxSize = MaxSize
	}
}
func SetMaxBackups(MaxBackups int) ModOptions {
	return func(option *Option) {
		option.MaxBackups = MaxBackups
	}
}
func SetMaxAge(MaxAge int) ModOptions {
	return func(option *Option) {
		option.MaxAge = MaxAge
	}
}

func SetLogFileDir(LogFileDir string) ModOptions {
	return func(option *Option) {
		option.LogDir = LogFileDir
	}
}

func SetFileName(FileName string) ModOptions {
	return func(option *Option) {
		option.FiLeName = FileName
	}
}

func SetLevel(Level zapcore.Level) ModOptions {
	return func(option *Option) {
		option.Level = Level
	}
}

func SetDevelopment(Development bool) ModOptions {
	return func(option *Option) {
		option.Development = Development
	}
}
