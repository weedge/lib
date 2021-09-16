// use function
package log

func Info(params ...interface{}) {
	DefaultLogger.Info(params)
}
func Debug(params ...interface{}) {
	DefaultLogger.Debug(params)
}
func Error(params ...interface{}) {
	DefaultLogger.Error(params)
}
