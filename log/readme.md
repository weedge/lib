#### 介绍

日志库，依赖 [zap](https://github.com/uber-go/zap)

#### todo

- [ ] 日志分级：基础日志 main(info,debug,warn,err,fatal),panic; 业务日志 biz; 访问请求日志 access,rpc; 启动日志
- [ ] 日志单元unit(k/v)记录,一次log输出
- [ ] 日志切分
- [ ] 自定义编码输出日志格式(zap自带 console ，json 两种日志，其他日志格式需要自定义)
- [ ] gin/grpc 日志中间件(middleware)/拦截器(Interceptor)

