// info,debug,error log info

package log

import (
	"go.uber.org/zap"
)

type mainLog struct {
	logger *zap.SugaredLogger
}

func newMainLog(config LoggerConfig, rotateByHour bool) *mainLog {
	logger := initFileLogger(config.Path, config.MinLevel, config.AddCaller, rotateByHour).Sugar()
	return &mainLog{
		logger: logger,
	}
}

func (l *mainLog) Info(args ...interface{}) {
	l.logger.Info(args...)
}
func (l *mainLog) Debug(args ...interface{}) {
	l.logger.Debug(args...)
}
func (l *mainLog) Warn(args ...interface{}) {
	l.logger.Warn(args...)
}
func (l *mainLog) Error(args ...interface{}) {
	l.logger.Error(args...)
}
func (l *mainLog) Infof(format string, args ...interface{}) {
	l.logger.Infof(format, args...)
}
func (l *mainLog) Debugf(format string, args ...interface{}) {
	l.logger.Debugf(format, args...)
}
func (l *mainLog) Warnf(format string, args ...interface{}) {
	l.logger.Warnf(format, args...)
}
func (l *mainLog) Errorf(format string, args ...interface{}) {
	l.logger.Errorf(format, args...)
}
