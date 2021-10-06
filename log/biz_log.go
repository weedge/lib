package log

import (
	"go.uber.org/zap"
)

type bizLog struct {
	logger *zap.SugaredLogger
}

func newBizLogger(config LoggerConfig, rotateByHour bool) (*bizLog, error) {
	logger, err := createZapLogger(config, rotateByHour)
	if err != nil {
		return nil, err
	}
	return &bizLog{
		logger: logger.Sugar(),
	}, nil
}

func (l *bizLog) Sync() {
	_ = l.logger.Sync()
}

func (l *bizLog) Info(args ...interface{}) {
	l.logger.Info(args...)
}
func (l *bizLog) Debug(args ...interface{}) {
	l.logger.Debug(args...)
}
func (l *bizLog) Warn(args ...interface{}) {
	l.logger.Warn(args...)
}
func (l *bizLog) Error(args ...interface{}) {
	l.logger.Error(args...)
}
func (l *bizLog) Infof(format string, args ...interface{}) {
	l.logger.Infof(format, args...)
}
func (l *bizLog) Debugf(format string, args ...interface{}) {
	l.logger.Debugf(format, args...)
}
func (l *bizLog) Warnf(format string, args ...interface{}) {
	l.logger.Warnf(format, args...)
}
func (l *bizLog) Errorf(format string, args ...interface{}) {
	l.logger.Errorf(format, args...)
}
