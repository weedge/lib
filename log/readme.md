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
```





