package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"time"
)

func TimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

type consoleLog struct {
	logger *zap.SugaredLogger
}

func newConsoleLog() *consoleLog {
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
			// Keys can be anything except the empty string.
			TimeKey:        "T",
			LevelKey:       "L",
			NameKey:        "N",
			CallerKey:      "C",
			MessageKey:     "M",
			StacktraceKey:  "S",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalLevelEncoder,
			EncodeTime:     TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		}),
		zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout)),
		zap.DebugLevel,
	)
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(2)).Sugar()
	return &consoleLog{
		logger: logger,
	}
}

func (l *consoleLog) Info(args ...interface{}) {
	l.logger.Info(args...)
}
func (l *consoleLog) Debug(args ...interface{}) {
	l.logger.Debug(args...)
}
func (l *consoleLog) Warn(args ...interface{}) {
	l.logger.Warn(args...)
}
func (l *consoleLog) Error(args ...interface{}) {
	l.logger.Error(args...)
}
func (l *consoleLog) Infof(format string, args ...interface{}) {
	l.logger.Infof(format, args...)
}
func (l *consoleLog) Debugf(format string, args ...interface{}) {
	l.logger.Debugf(format, args...)
}
func (l *consoleLog) Warnf(format string, args ...interface{}) {
	l.logger.Warnf(format, args...)
}
func (l *consoleLog) Errorf(format string, args ...interface{}) {
	l.logger.Errorf(format, args...)
}
