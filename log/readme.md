#### 介绍

日志库，依赖三方日志库:  uber [zap](https://github.com/uber-go/zap) , 日志按时间切割[lestrrat-go/file-rotatelogs](github.com/lestrrat-go/file-rotatelogs)；不设置日志目录路径，默认打印console日志。

#### 功能:

1. 日志分级：基础日志 main(info,debug,warn,err,fatal),panic; 业务日志 biz; 访问请求日志 access,rpc; 启动日志
2. 日志单元unit(k/v)记录,一次log输出
3. 日志切分, 每小时切分一次
4. 自定义编码输出日志格式encoder(zap自带 console ，json 两种日志，其他日志格式需要自定义,参考json日志encoder)
5. gin/grpc 访问日志中间件(middleware)

#### 接口：

```go
// setup log with those params:
// project name for tapper log
// log.json config path,
// default log path for log.json undefined log path
func Setup(projectName string, confPath string, defaultLogPath string) error 

AccessInfo(msg string, fields ...zap.Field)

Info(args ...interface{})
Debug(args ...interface{})
Warn(args ...interface{})
Error(args ...interface{})
Infof(format string, args ...interface{})
Debugf(format string, args ...interface{})
Warnf(format string, args ...interface{})
Errorf(format string, args ...interface{})
func RpcInfo(params ...interface{})
func RpcInfof(format string, params ...interface{})
func Recover(v ...interface{})
func Recoverf(format string, params ...interface{})

// flush main, biz, access, panic, rpc log
// Sync flushes any buffered log entries.
func FlushLog() 

```

测试见example_test

#### 配置 (log.json)

```json
{
  "logs": [
    {
      "logger": "main",// log type:main(debug,info,warn)log, err log, access log, biz log, panic log, rpc log 
      "min_level": "debug",// log min level 
      "add_caller": true,// zapcore addCaller open, skip to show caller line
      "policy": "filter",// filter  zapTee 
      "filters": [
        {
          "level": "debug,info,warn",// log level
          "path": "./log/zap.log" // log path
        },
        {
          "level": "error",
          "path": "./log/zap.err.log"
        }
      ]
    },
    {
      "logger": "access",
      "min_level": "info",
      "policy": "file",
      "path": "./log/zap-access.log"
    },
    {
      "logger": "biz",
      "min_level": "info",
      "add_caller": true,
      "policy": "file",
      "path": "./log/zap-biz.log"
    },
    {
      "logger": "panic",
      "min_level": "info",
      "add_caller": true,
      "policy": "file",
      "path": "./log/zap-reco.log"
    },
    {
      "logger": "rpc",
      "min_level": "info",
      "add_caller": true,
      "policy": "file",
      "path": "./log/zap-rpc.log"
    }
  ],
  "rotateByHour": true //open rotate log per hour:00, if deploy to docker in k8s, close
}

```



