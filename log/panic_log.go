// panic log info

package log

import (
	"go.uber.org/zap"
)

type panicLog struct {
	logger *zap.SugaredLogger
}

func newPanicLogger(config LoggerConfig, rotateByHour bool) (*panicLog, error) {
	logger, err := createZapLogger(config, rotateByHour)
	if err != nil {
		return nil, err
	}
	return &panicLog{
		logger: logger.Sugar(),
	}, nil
}

func (l *panicLog) Sync() {
	_ = l.logger.Sync()
}

func (l *panicLog) Info(args ...interface{}) {
	l.logger.Info(args...)
}
func (l *panicLog) Debug(args ...interface{}) {
	l.logger.Debug(args...)
}
func (l *panicLog) Warn(args ...interface{}) {
	l.logger.Warn(args...)
}
func (l *panicLog) Error(args ...interface{}) {
	l.logger.Error(args...)
}
func (l *panicLog) Infof(format string, args ...interface{}) {
	l.logger.Infof(format, args...)
}
func (l *panicLog) Debugf(format string, args ...interface{}) {
	l.logger.Debugf(format, args...)
}
func (l *panicLog) Warnf(format string, args ...interface{}) {
	l.logger.Warnf(format, args...)
}
func (l *panicLog) Errorf(format string, args ...interface{}) {
	l.logger.Errorf(format, args...)
}

func (l *panicLog) FatalF(format string, args ...interface{}) {
	l.logger.Fatalf(format, args...)
}
func (l *panicLog) Fatal(args ...interface{}) {
	l.logger.Fatal(args...)
}
