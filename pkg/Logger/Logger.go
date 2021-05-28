package Logger

import (
	"fmt"
	"go.uber.org/zap"
)

func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.log.Debug(msg, fields...)
}

func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.log.Info(msg, fields...)
}

func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.log.Warn(msg, fields...)
}

func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.log.Error(msg, fields...)
}

func (l *Logger) Panic(msg string, fields ...zap.Field) {
	l.log.Panic(msg, fields...)
}

func (l *Logger) Debugf(template string, args ...interface{}) {
	l.log.Debug(sprintf(template, args...))
}

func (l *Logger) Infof(template string, args ...interface{}) {
	l.log.Info(sprintf(template, args...))
}

func (l *Logger) Warnf(template string, args ...interface{}) {
	l.log.Warn(sprintf(template, args...))
}

func (l *Logger) Errorf(template string, args ...interface{}) {
	l.log.Error(sprintf(template, args...))
}

func (l *Logger) Panicf(template string, args ...interface{}) {
	l.log.Panic(sprintf(template, args...))
}

func sprintf(template string, args ...interface{}) string {
	msg := template
	if template == "" && len(args) > 0 {
		msg = fmt.Sprint(args...)
	} else if msg != "" && len(args) > 0 {
		msg = fmt.Sprintf(msg, args)
	}
	return msg
}
