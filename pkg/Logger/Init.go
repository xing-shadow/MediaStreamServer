package Logger

import (
	"fmt"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path/filepath"
	"time"
)

var (
	logger                         = new(Logger)
	opt                            = new(Option)
	errWS, warnWS, infoWS, debugWS zapcore.WriteSyncer
	debugConsoleWS                 = zapcore.Lock(os.Stdout)
	errorConsoleWS                 = zapcore.Lock(os.Stderr)
)

func GetLogger() *zap.Logger {
	return logger.log
}

func Init(opts ...ModOptions) (err error) {
	if logger.inited {
		logger.log.Info("[NewLogger] logger Inited")
		return nil
	}
	for _, item := range opts {
		item(opt)
	}
	opt.fixup()
	logger.Opt = opt
	if opt.Development {
		logger.zapConfig = zap.NewDevelopmentConfig()
	} else {
		logger.zapConfig = zap.NewProductionConfig()
	}
	{
		logger.zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
		logger.zapConfig.EncoderConfig.EncodeTime = timeEncoder
	}
	logger.zapConfig.DisableStacktrace = true
	logger.zapConfig.Level.SetLevel(opt.Level)
	logger.inited = true
	if err = logger.init(); err != nil {
		return
	}
	fmt.Println("[NewLogger] success")
	return nil
}

type Logger struct {
	Opt       *Option
	inited    bool
	log       *zap.Logger
	zapConfig zap.Config
}

func (l *Logger) init() error {
	l.setSyncers()
	var err error
	l.log, err = l.zapConfig.Build(l.cors())
	if err != nil {
		return err
	}
	return nil
}

func (l *Logger) setSyncers() {
	f := func(filePath string) zapcore.WriteSyncer {
		return zapcore.AddSync(&lumberjack.Logger{
			Filename:   filePath,
			MaxSize:    l.Opt.MaxSize,
			MaxAge:     l.Opt.MaxAge,
			MaxBackups: l.Opt.MaxBackups,
			LocalTime:  true,
			Compress:   false,
		})
	}
	debugFilePath := filepath.Join(opt.LogDir, fmt.Sprintf("%s-debug.log", opt.FiLeName))
	debugWS = f(debugFilePath)
	infoFilePath := filepath.Join(opt.LogDir, fmt.Sprintf("%s-info.log", opt.FiLeName))
	infoWS = f(infoFilePath)
	warnFilePath := filepath.Join(opt.LogDir, fmt.Sprintf("%s-warn.log", opt.FiLeName))
	warnWS = f(warnFilePath)
	errorFilePath := filepath.Join(opt.LogDir, fmt.Sprintf("%s-error.log", opt.FiLeName))
	errWS = f(errorFilePath)
}

func (l *Logger) cors() zap.Option {
	fileEncoder := zapcore.NewJSONEncoder(l.zapConfig.EncoderConfig)
	encoderConfig := zap.NewDevelopmentEncoderConfig()
	encoderConfig.EncodeTime = timeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
	errPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl == zapcore.ErrorLevel && zapcore.ErrorLevel-l.zapConfig.Level.Level() > -1
	})
	warnPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl == zapcore.WarnLevel && zapcore.WarnLevel-l.zapConfig.Level.Level() > -1
	})
	infoPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl == zapcore.InfoLevel && zapcore.InfoLevel-l.zapConfig.Level.Level() > -1
	})
	debugPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl == zapcore.DebugLevel && zapcore.DebugLevel-l.zapConfig.Level.Level() > -1
	})
	var cores []zapcore.Core
	if l.Opt.Development {
		cores = append(cores, []zapcore.Core{
			zapcore.NewCore(consoleEncoder, errorConsoleWS, errPriority),
			zapcore.NewCore(consoleEncoder, debugConsoleWS, warnPriority),
			zapcore.NewCore(consoleEncoder, debugConsoleWS, infoPriority),
			zapcore.NewCore(consoleEncoder, debugConsoleWS, debugPriority),
		}...)
	} else {
		cores = append(cores, []zapcore.Core{
			zapcore.NewCore(fileEncoder, errWS, errPriority),
			zapcore.NewCore(fileEncoder, warnWS, warnPriority),
			zapcore.NewCore(fileEncoder, infoWS, infoPriority),
			zapcore.NewCore(fileEncoder, debugWS, debugPriority),
		}...)
	}
	return zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return zapcore.NewTee(cores...)
	})
}

func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05"))
}
