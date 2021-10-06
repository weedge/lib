// use function
package log

import (
	"fmt"
	"go.uber.org/zap"

	"path/filepath"
)

// setup log with those params:
// project name for tapper log
// log.json config path,
// default log path for log.json undefined log path
func Setup(projectName string, confPath string, defaultLogPath string) error {
	logConfPath := filepath.Join(confPath, defaultLogConfFileName)
	config, err := getLogConfig(logConfPath)
	if err != nil {
		return err
	}

	return createLogger(config, projectName, defaultLogPath)
}

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
	if mainLogger != nil {
		mainLogger.Error(params...)
		return
	}
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
	if mainLogger != nil {
		mainLogger.Errorf(format, params...)
		return
	}
	defaultLogger.Errorf(format, params...)
}

func BizArchive(format string, params ...interface{}) {
	if bizLogger != nil {
		bizLogger.Infof(format, params...)
		return
	}
	defaultLogger.Infof(format, params...)
}

func AccessInfo(msg string, fields ...zap.Field) {
	if accessLogger != nil {
		accessLogger.Info(msg, fields...)
		return
	}
	defaultLogger.AccessInfo(msg, fields...)
}

func RpcInfo(params ...interface{}) {
	if rpcLogger != nil {
		rpcLogger.Info(params...)
		return
	}
	defaultLogger.Info(params...)
}

func RpcInfof(format string, params ...interface{}) {
	if rpcLogger != nil {
		rpcLogger.Infof(format, params...)
		return
	}
	defaultLogger.Infof(format, params...)
}

// recover panic log
func Recover(v ...interface{}) {
	if panicLogger != nil {
		//panicLogger.Fatal(fmt.Sprint(v...))
		panicLogger.Error(fmt.Sprint(v...))
		return
	}
	//defaultLogger.Fatal(fmt.Sprint(v...))
	defaultLogger.Error(fmt.Sprint(v...))
}

// recover panic format log
func Recoverf(format string, params ...interface{}) {
	if panicLogger != nil {
		//panicLogger.Fatal(fmt.Sprint(params...))
		panicLogger.Error(fmt.Sprint(params...))
		return
	}
	//defaultLogger.Fatal(fmt.Sprint(params...))
	defaultLogger.Error(fmt.Sprint(params...))
}

// flush main, biz, access, panic, rpc log
// Sync flushes any buffered log entries.
func FlushLog() {
	if mainLogger != nil {
		mainLogger.Sync()
	}
	if bizLogger != nil {
		bizLogger.Sync()
	}
	if accessLogger != nil {
		accessLogger.Sync()
	}
	if panicLogger != nil {
		panicLogger.Sync()
	}
	if rpcLogger != nil {
		rpcLogger.Sync()
	}
}

