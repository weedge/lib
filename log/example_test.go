package log

import (
	"go.uber.org/zap"
)

func Example_Info() {
	Info("123123", "adfasdf")
	Debug("123123", "adfasdf")
	Warn("123123", "adfasdf")
	Error("123123", "adfasdf")
	RpcInfo("123123", "rpc")
	Recover("123123", "recover")

	// output:
	//
}

func Example_Infof() {
	Infof("%s, %s", "123123", "adfasdf")
	Debugf("%s, %s", "123123", "adfasdf")
	Warnf("%s, %s", "123123", "adfasdf")
	Errorf("%s, %s", "123123", "adfasdf")
	BizArchive("%s, %s", "123123", "biz")
	RpcInfof("%s, %s", "123123", "rpc")
	Recoverf("%s, %s", "123123", "recover")

	// output:
	//
}

func ExampleAccessInfo() {
	fields := []zap.Field{
		zap.String(LOGID, "123123"),
		zap.String(URI, "/source/get"),
		zap.String(REFERER, "www.google.com"),
		zap.Any(COOKIE, "cookie:1231"),
		zap.String(CLIENT_IP, "192.168.7.3"),
		zap.String(LOCAL_IP, LOCALIP),
		zap.String(UA, "web-chrome"),
		zap.String(HOST, "www.baidu.com"),
		zap.String(MODULE, "live"),
		zap.String("request_param", "a=1&b=2"),
		zap.String(NOTICES, "notices"),
		zap.String(MONITOR, "{'monitor':true}"),
		zap.Int(CODE, 200),
		zap.String(RESPONSE, "response"),
		zap.Int64(COST, 100),
	}
	AccessInfo("", fields...)

	// output:
	//
}

func ExampleSetup() {
	err := Setup("testProject", "./", "./log", nil)
	if err != nil {
		println(err)
	}
	Example_Info()
	Example_Infof()
	ExampleAccessInfo()

	// output:
	//
}
