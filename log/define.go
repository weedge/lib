package log

import (
	"path/filepath"

	"github.com/weedge/lib/log/tapper"
)

var (
	defaultLogger = newConsoleLogger()

	mainLogger   *mainLog
	bizLogger    *bizLog
	panicLogger  *panicLog
	accessLogger *accessLog
	rpcLogger    *rpcLog
)

var (
	defaultLogConfFileName = "log.json"
)

func createLogger(conf *LogConfig, projectName, defaultLogPath string) (err error) {
	tapper.Project = projectName
	for _, loggerConf := range conf.Logs {
		if loggerConf.Path == "" {
			loggerConf.Path = filepath.Join(defaultLogPath, projectName+"-"+loggerConf.Logger+".log")
		}
		if loggerConf.MinLevel == "" {
			loggerConf.MinLevel = "info"
		}
		if loggerConf.Policy == "" {
			loggerConf.Policy = "file"
		}
		switch loggerConf.Logger {
		case "main":
			mainLogger, err = newMainLogger(loggerConf, conf.RotateByHour)
		case "biz":
			bizLogger, err = newBizLogger(loggerConf, conf.RotateByHour)
		case "panic":
			panicLogger, err = newPanicLogger(loggerConf, conf.RotateByHour)
		case "access":
			accessLogger, err = newAccessLogger(loggerConf, conf.RotateByHour)
		case "rpc":
			rpcLogger, err = newRpcLogger(loggerConf, conf.RotateByHour)
		}
	}

	return
}
