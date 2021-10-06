package log

import (
	"os"
	"time"

	"github.com/weedge/lib/log/encoder"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

type consoleLog struct {
	logger *zap.Logger
}

func newConsoleLogger() *consoleLog {
	core := zapcore.NewCore(
		encoder.NewSelfEncoder(zapcore.EncoderConfig{
			//zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
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
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(2))
	return &consoleLog{
		logger: logger,
	}
}

func (l *consoleLog) AccessInfo(msg string, fields ...zap.Field) {
	l.logger.Info(msg, fields...)
}

func (l *consoleLog) Info(args ...interface{}) {
	l.logger.Sugar().Info(args...)
}
func (l *consoleLog) Debug(args ...interface{}) {
	l.logger.Sugar().Debug(args...)
}
func (l *consoleLog) Warn(args ...interface{}) {
	l.logger.Sugar().Warn(args...)
}
func (l *consoleLog) Error(args ...interface{}) {
	l.logger.Sugar().Error(args...)
}
func (l *consoleLog) Infof(format string, args ...interface{}) {
	l.logger.Sugar().Infof(format, args...)
}
func (l *consoleLog) Debugf(format string, args ...interface{}) {
	l.logger.Sugar().Debugf(format, args...)
}
func (l *consoleLog) Warnf(format string, args ...interface{}) {
	l.logger.Sugar().Warnf(format, args...)
}
func (l *consoleLog) Errorf(format string, args ...interface{}) {
	l.logger.Sugar().Errorf(format, args...)
}
func (l *consoleLog) FatalF(format string, args ...interface{}) {
	l.logger.Sugar().Fatalf(format, args...)
}
func (l *consoleLog) Fatal(args ...interface{}) {
	l.logger.Sugar().Fatal(args...)
}
