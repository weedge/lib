package log

import (
	"go.uber.org/zap"
)

type accessLog struct {
	logger *zap.Logger
}

func newAccessLogger(config LoggerConfig, rotateByHour bool) (*accessLog, error) {
	logger, err := createZapLogger(config, rotateByHour)
	if err != nil {
		return nil, err
	}
	return &accessLog{
		logger: logger,
	}, err
}

func (l *accessLog) Sync() {
	_ = l.logger.Sync()
}

func (l *accessLog) Info(msg string, fields ...zap.Field) {
	l.logger.Info(msg, fields...)
}
