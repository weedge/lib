// use function
package log

func Info(params ...interface{}) {
	DefaultLogger.Info(params)
}
func Debug(params ...interface{}) {
	DefaultLogger.Debug(params)
}
func Warn(params ...interface{}) {
	DefaultLogger.Warn(params)
}
func Error(params ...interface{}) {
	DefaultLogger.Error(params)
}
func Infof(format string, params ...interface{}) {
	DefaultLogger.Infof(format, params)
}
func Debugf(format string, params ...interface{}) {
	DefaultLogger.Debugf(format, params)
}
func Warnf(format string, params ...interface{}) {
	DefaultLogger.Warnf(format, params)
}
func Errorf(format string, params ...interface{}) {
	DefaultLogger.Errorf(format, params)
}
