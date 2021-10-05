package log

var (
	defaultLogger logger = newConsoleLog()

	mainLogger    logger
	bizLogger     logger
	accessLogger  logger
	recoverLogger logger
	ralLogger     logger
)

type logger interface {
	Info(args ...interface{})
	Debug(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Infof(format string, args ...interface{})
	Debugf(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

/*
func SetUp(){
	mainLogger = newMainLog()
}

*/
