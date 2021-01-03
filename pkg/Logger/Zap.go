package Logger

import (
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Options struct {
	LogFileDir string
	AppName    string
	MaxSize    int //文件多大开始切分
	MaxBackups int //保留文件个数
	MaxAge     int //文件保留最大实际
	Level      string
}

var (
	l              *Logger
	sp             = string(filepath.Separator)
	outWrite       zapcore.WriteSyncer       // IO输出
	debugConsoleWS = zapcore.Lock(os.Stdout) // 控制台标准输出
)

func init() {
	l = &Logger{
		Opts: &Options{
			Level:      "debug",
		},
	}
	NewLogger()
}

type Logger struct {
	*zap.Logger
	sync.RWMutex
	Opts      *Options
	zapConfig zap.Config
	inited    bool
}

func NewLogger(cf ...*Options) {
	l.Lock()
	defer l.Unlock()
	if l.inited {
		l.Info("[initLogger] logger Inited")
		return
	}
	if len(cf) > 0 {
		l.Opts = cf[0]
	}
	l.loadCfg()
	l.init()
	l.inited = true
}

// GetLogger returns logger
func GetLogger() (ret *Logger) {
	return l
}

func (l *Logger) init() {
	l.setSyncers()
	var err error
	l.Logger, err = l.zapConfig.Build(l.cores())
	if err != nil {
		panic(err)
	}
	defer l.Logger.Sync()
}
func (l *Logger) GetLevel() (level zapcore.Level) {
	switch strings.ToLower(l.Opts.Level) {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "dpanic":
		return zapcore.DPanicLevel
	case "panic":
		return zapcore.PanicLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.DebugLevel //默认为调试模式
	}
}

func (l *Logger) loadCfg() {
	if l.GetLevel() == zapcore.DebugLevel {
		l.zapConfig = zap.NewDevelopmentConfig()
	} else {
		l.zapConfig = zap.NewProductionConfig()
	}
	l.zapConfig.OutputPaths = []string{"stdout"}
	l.zapConfig.ErrorOutputPaths = []string{"stderr"}
	// 默认输出到程序运行目录的logs子目录
	if l.Opts.LogFileDir == "" {
		l.Opts.LogFileDir, _ = filepath.Abs(filepath.Dir(filepath.Join(".")))
		l.Opts.LogFileDir += sp + "logs" + sp
	}
	if l.Opts.AppName == "" {
		appName := filepath.Base(os.Args[0])
		l.Opts.AppName = appName
	}
	if l.Opts.MaxSize == 0 {
		l.Opts.MaxSize = 100
	}
	if l.Opts.MaxBackups == 0 {
		l.Opts.MaxBackups = 60
	}
	if l.Opts.MaxAge == 0 {
		l.Opts.MaxAge = 30
	}
}

func (l *Logger) setSyncers() {
	outWrite = zapcore.AddSync(&lumberjack.Logger{
		Filename:   l.Opts.LogFileDir + sp + l.Opts.AppName + ".log",
		MaxSize:    l.Opts.MaxSize,
		MaxBackups: l.Opts.MaxBackups,
		MaxAge:     l.Opts.MaxAge,
		Compress:   true,
		LocalTime:  true,
	})
	return
}

func (l *Logger) cores() zap.Option {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "Time",
		LevelKey:       "Level",
		NameKey:        "logger",
		CallerKey:      "Caller",
		MessageKey:     "Msg",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder, //这里可以指定颜色
		EncodeTime:     timeEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,      // 短路径编码器
	}
	fileEncoder := NewCustomJsonEncoder(encoderConfig)
	priority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= l.GetLevel()
	})
	var cores []zapcore.Core
	if l.GetLevel() == zapcore.DebugLevel {
		cores = append(cores, []zapcore.Core{
			zapcore.NewCore(fileEncoder, debugConsoleWS, priority),
		}...)
	}else {
		cores = append(cores, []zapcore.Core{
			zapcore.NewCore(fileEncoder, outWrite, priority),
		}...)
	}

	return zap.WrapCore(func(c zapcore.Core) zapcore.Core {
		return zapcore.NewTee(cores...)
	})
}
func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05"))
}
