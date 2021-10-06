package log

import (
	"go.uber.org/zap"
)

type rpcLog struct {
	logger *zap.SugaredLogger
}

func newRpcLogger(config LoggerConfig, rotateByHour bool) (*rpcLog, error) {
	logger, err := createZapLogger(config, rotateByHour)
	if err != nil {
		return nil, err
	}
	return &rpcLog{
		logger: logger.Sugar(),
	}, nil
}

func (l *rpcLog) Sync() {
	_ = l.logger.Sync()
}

func (l *rpcLog) Info(args ...interface{}) {
	l.logger.Info(args...)
}
func (l *rpcLog) Debug(args ...interface{}) {
	l.logger.Debug(args...)
}
func (l *rpcLog) Warn(args ...interface{}) {
	l.logger.Warn(args...)
}
func (l *rpcLog) Error(args ...interface{}) {
	l.logger.Error(args...)
}
func (l *rpcLog) Infof(format string, args ...interface{}) {
	l.logger.Infof(format, args...)
}
func (l *rpcLog) Debugf(format string, args ...interface{}) {
	l.logger.Debugf(format, args...)
}
func (l *rpcLog) Warnf(format string, args ...interface{}) {
	l.logger.Warnf(format, args...)
}
func (l *rpcLog) Errorf(format string, args ...interface{}) {
	l.logger.Errorf(format, args...)
}
