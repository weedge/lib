// use function
package log

func Info(params ...interface{}) {
	if mainLogger != nil {
		mainLogger.Info(params...)
		return
	}
	defaultLogger.Info(params...)
}
func Debug(params ...interface{}) {
	if mainLogger != nil {
		mainLogger.Debug(params...)
		return
	}
	defaultLogger.Debug(params...)
}
func Warn(params ...interface{}) {
	if mainLogger != nil {
		mainLogger.Warn(params...)
		return
	}
	defaultLogger.Warn(params...)
}
func Error(params ...interface{}) {
	defaultLogger.Error(params...)
}
func Infof(format string, params ...interface{}) {
	if mainLogger != nil {
		mainLogger.Infof(format, params...)
		return
	}
	defaultLogger.Infof(format, params...)
}
func Debugf(format string, params ...interface{}) {
	if mainLogger != nil {
		mainLogger.Debugf(format, params...)
		return
	}
	defaultLogger.Debugf(format, params...)
}
func Warnf(format string, params ...interface{}) {
	if mainLogger != nil {
		mainLogger.Warnf(format, params...)
		return
	}
	defaultLogger.Warnf(format, params...)
}
func Errorf(format string, params ...interface{}) {
	defaultLogger.Errorf(format, params...)
}
