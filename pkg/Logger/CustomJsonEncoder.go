package Logger

import (
	"fmt"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
	"strings"
)
const DefaultLineEnding = "\n"

var (
	_pool = buffer.NewPool()
	// Get retrieves a buffer from the pool, creating one if necessary.
	Get = _pool.Get
)

type CustomJsonEncoder struct {
	CustomInterface
	*zapcore.EncoderConfig
}

type CustomInterface interface {
	zapcore.ObjectEncoder
	Clone() zapcore.Encoder
}

func NewCustomJsonEncoder(cfg zapcore.EncoderConfig) zapcore.Encoder {
	if cfg.ConsoleSeparator == "" {

	}
	return &CustomJsonEncoder{
		CustomInterface: zapcore.NewConsoleEncoder(cfg),
		EncoderConfig:   &cfg,
	}
}

func (c CustomJsonEncoder) addSeparatorIfNecessary(line *buffer.Buffer) {
	if line.Len() > 0 {
		line.AppendString(c.ConsoleSeparator)
	}
}

func (c *CustomJsonEncoder) EncodeEntry(ent zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	line := Get()
	//Level
	if ent.Level.String() != ""  {
		switch ent.Level {
		case zapcore.DebugLevel:
			line.AppendString(fmt.Sprintf("\x1b[%dm%s\x1b[0m",Magenta,strings.ToUpper(ent.Level.String())))
		case zapcore.InfoLevel:
			line.AppendString(fmt.Sprintf("\x1b[%dm%s\x1b[0m",Cyan,strings.ToUpper(ent.Level.String())))
		case zapcore.WarnLevel:
			line.AppendString(fmt.Sprintf("\x1b[%dm%s\x1b[0m",Yellow,strings.ToUpper(ent.Level.String())))
		case zapcore.PanicLevel:
			fallthrough
		case zapcore.FatalLevel:
			fallthrough
		case zapcore.ErrorLevel:
			line.AppendString(fmt.Sprintf("\x1b[%dm%s\x1b[0m",Red,strings.ToUpper(ent.Level.String())))
		}
		line.AppendByte(9)
	}
	//Timec
	if c.TimeKey != "" && c.EncodeTime != nil {
		line.AppendString(ent.Time.Format("2006-01-02T15:04:05"))
		line.AppendByte(9)
	}
	//Caller
	if ent.Caller.Defined {
		line.AppendString(ent.Caller.TrimmedPath())
		line.AppendByte(9)
	}
	//files
	if len(fields) != 0{
		line.AppendByte('<')
		for _, field := range fields {
			line.AppendString(fmt.Sprintf("%s:%s ",field.Key,field.String))
		}
		line.AppendByte('>')
	}
	// Message
	if c.MessageKey != "" {
		line.AppendString(DefaultLineEnding)
		line.AppendString(fmt.Sprintf("    %s",ent.Message))
	}

	line.AppendString(DefaultLineEnding)
	return line, nil
}

func FullNameEncoder(loggerName string, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(loggerName)
}
