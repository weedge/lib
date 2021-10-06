package log

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/weedge/lib/log/encoder"

	"github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func createZapLogger(config LoggerConfig, rotateByHour bool) (*zap.Logger, error) {
	switch strings.ToLower(config.Policy) {
	case "file":
		if len(config.Path) <= 0 {
			return nil, errors.New("empty path")
		}
		return initFileLogger(config.Path, config.MinLevel, config.AddCaller, rotateByHour), nil
	case "filter":
		if len(config.Filters) <= 0 {
			return nil, errors.New("empty filters")
		}
		return initFiltersLogger(config.Filters, config.MinLevel, config.AddCaller, rotateByHour), nil
	default:
		return nil, errors.New("invalid policy")
	}
}

func createDefaultLogger(defaultLogPath string, filename string, addCaller, rotateByHour bool) *zap.Logger {
	logPath := filepath.Join(defaultLogPath, filename)
	return initFileLogger(logPath, "info", addCaller, rotateByHour)
}

func initFiltersLogger(filters []Filter, minLevel string, addCaller, rotateByHour bool) *zap.Logger {
	minLvl := getLevel(minLevel)
	encoderConfig := getDefaultEncoderConfig()
	cores := make([]zapcore.Core, 0, len(filters))
	for _, filter := range filters {
		w := getWriter(filter.Path, rotateByHour)
		filterLevels := getLevels(filter.Level)
		allows := make(map[zapcore.Level]bool)
		for _, filterLevel := range filterLevels {
			allows[filterLevel] = true
		}
		levelEnabler := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl >= minLvl && allows[lvl]
		})
		core := zapcore.NewCore(encoder.NewSelfEncoder(encoderConfig), w, levelEnabler)
		cores = append(cores, core)
	}
	core := zapcore.NewTee(cores...)
	if addCaller {
		return zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	} else {
		return zap.New(core)
	}
}

func initFileLogger(logPath string, minLevel string, addCaller, rotateByHour bool) *zap.Logger {
	encoderConfig := getDefaultEncoderConfig()
	w := getWriter(logPath, rotateByHour)
	level := getLevel(minLevel)
	core := zapcore.NewCore(encoder.NewSelfEncoder(encoderConfig), w, level)
	if addCaller {
		return zap.New(core, zap.AddCaller(), zap.AddCallerSkip(2))
	} else {
		return zap.New(core)
	}
}

func getDefaultEncoderConfig() zapcore.EncoderConfig {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05"))
	}
	return encoderConfig
}

func getWriter(logPath string, rotateByHour bool) zapcore.WriteSyncer {
	if err := createLogDir(logPath); err != nil {
		panic(err)
	}

	/*
		logger, err := os.Create(logPath)
		if err != nil {
			panic(err)
		}
	*/
	if rotateByHour == false {
		logger := &Logger{
			Filename: logPath,
			mu:       sync.Mutex{},
		}

		return zapcore.AddSync(logger)
	}
	// save 14 days log, rotate log per hour::00
	hook, err := rotatelogs.New(
		logPath+".%Y%m%d%H",
		rotatelogs.WithMaxAge(time.Hour*24*14),
		rotatelogs.WithRotationTime(time.Hour),
	)
	if err != nil {
		panic(err)
	}
	return zapcore.AddSync(hook)
}

func createLogDir(logPath string) error {
	if dir := filepath.Dir(logPath); len(dir) > 0 {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return fmt.Errorf("can't make directories for new logfile: %s", err.Error())
		}
	}
	return nil
}

func getLevels(levels string) []zapcore.Level {
	ls := strings.Split(levels, ",")
	res := make([]zapcore.Level, 0, len(ls))
	for _, l := range ls {
		res = append(res, getLevel(l))
	}
	return res
}

func getLevel(level string) zapcore.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zap.DebugLevel
	case "info":
		return zap.InfoLevel
	case "warn", "warning":
		return zap.WarnLevel
	case "error":
		return zap.ErrorLevel
	case "panic", "dpanic":
		return zap.PanicLevel
	case "fatal":
		return zap.FatalLevel
	default:
		return zap.InfoLevel
	}
}
