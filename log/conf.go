package log

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
)

type LogConfig struct {
	Logs         []LoggerConfig `json:"logs"`
	RotateByHour bool           `json:"rotateByHour"`
}

type LoggerConfig struct {
	Logger    string   `json:"logger"`
	MinLevel  string   `json:"min_level"`
	AddCaller bool     `json:"add_caller"`
	Policy    string   `json:"policy"`
	Filters   []Filter `json:"filters"`
	Path      string   `json:"path"`
}

type Filter struct {
	Level string `json:"level"`
	Path  string `json:"path"`
}

func IsExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

func LoggerConfigAsFile(filename string) (*LogConfig, error) {
	if !IsExist(filename) {
		return &LogConfig{}, nil
	}
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	conf, err := configFromReader(file)
	if err != nil {
		return nil, err
	}
	return conf, nil
}

func configFromReader(reader io.Reader) (*LogConfig, error) {
	return unmarshalConfig(reader)
}

func unmarshalConfig(reader io.Reader) (*LogConfig, error) {
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	logConfig := &LogConfig{}
	err = json.Unmarshal(bytes, logConfig)
	if err != nil {
		return nil, err
	}
	return logConfig, nil
}
