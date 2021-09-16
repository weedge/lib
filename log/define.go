package log

var DefaultLogger logger = newDefaultLog()

type logger interface {
	Error(args ...interface{})
	Info(args ...interface{})
	Debug(args ...interface{})
	Errorf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Debugf(format string, args ...interface{})
}
