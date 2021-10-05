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

type defaultLog struct {
	logger *zap.SugaredLogger
}

func newDefaultLog() *defaultLog {
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
	return &defaultLog{
		logger: logger,
	}
}

func (l *defaultLog) Info(args ...interface{}) {
	l.logger.Info(args...)
}
func (l *defaultLog) Debug(args ...interface{}) {
	l.logger.Debug(args...)
}
func (l *defaultLog) Warn(args ...interface{}) {
	l.logger.Warn(args...)
}
func (l *defaultLog) Error(args ...interface{}) {
	l.logger.Error(args...)
}
func (l *defaultLog) Infof(format string, args ...interface{}) {
	l.logger.Infof(format, args...)
}
func (l *defaultLog) Debugf(format string, args ...interface{}) {
	l.logger.Debugf(format, args...)
}
func (l *defaultLog) Warnf(format string, args ...interface{}) {
	l.logger.Warnf(format, args...)
}
func (l *defaultLog) Errorf(format string, args ...interface{}) {
	l.logger.Errorf(format, args...)
}
