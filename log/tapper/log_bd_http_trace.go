package tapper

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	/*
		X_bd_caller_uri: 3284,/icourse/main/index,0,0,/icourse,/icourse
		X_bd_idc: yun
		X_bd_logid: 32541783
		X_bd_module: icourse
		X_bd_product: homework
		X_bd_spanid: 0.3
		X_bd_subsys: question
		X_bd_uniqid: 3528580869
		X_bd_userip: 183.93.17.119
	*/

	// 定义spanContext
	HTTP_KEY_TRACE_ID = "X_bd_logid"
	HTTP_KEY_SPAN_ID  = "X_bd_spanid" // 0.0.0 规则, 此处第一调用为，0.3.0
	HTTP_KEY_UNIQ_ID  = "X_bd_uniqid"
	HTTP_KEY_CALLER   = "X_bd_module"

	HTTP_KEY_IDC        = "X_bd_idc"
	HTTP_KEY_CALLER_URI = "X_bd_caller_uri" //osc 限流，此处不设置
	HTTP_KEY_SUBSYS     = "X_bd_subsys"

	// 自定义baggage，需全程透传
	HTTP_KEY_PRODUCT = "X_bd_product"
	HTTP_KEY_USERIP  = "X_bd_userip"

	HTTP_USER_AGENT = "User-Agent"
)

var Project = ""
var _USER_AGENT = "ral/go "

// 生成logId
func GenLogId() string {
	now := time.Now()
	logId := ((now.UnixNano()*1e5 + now.UnixNano()/1e7) & 0x7FFFFFFF) | 0x80000000
	return strconv.FormatInt(logId, 10)
}

func SetTraceLogFromGinHeader(c *gin.Context) *TraceLog {
	traceLog := &TraceLog{}
	//获取当前LogId
	bdLogId := c.GetHeader(HTTP_KEY_TRACE_ID)
	if bdLogId != "" && bdLogId != "0" {
		traceLog.LogId = strings.TrimSpace(bdLogId)
	} else {
		traceLog.LogId = GenLogId()
	}
	//获取调用者
	traceLog.Caller = c.GetHeader(HTTP_KEY_CALLER)
	//获取调用者refer
	//traceLog.Refer = c.GetHeader(HTTP_KEY_CALLER_URI)
	//调用userip
	traceLog.UniqId = c.GetHeader(HTTP_KEY_UNIQ_ID)
	//获取当前的调用接口路径
	traceLog.Path = c.Request.URL.Path

	//设置mq的请求transid
	traceLog.MqTransId = c.Query("transid")

	//全程透传
	traceLog.UserIP = c.GetHeader(HTTP_KEY_USERIP)
	traceLog.Product = c.GetHeader(HTTP_KEY_PRODUCT)
	c.Set(TRACECTX, traceLog)
	return traceLog
}

func SetTraceHeaderByGinContext(ctx context.Context, header map[string]string) map[string]string {
	if header == nil {
		header = make(map[string]string)
	}
	traceLog, _ := GetTraceLogFromGinContext(ctx)

	header[HTTP_KEY_TRACE_ID] = traceLog.LogId
	header[HTTP_KEY_SPAN_ID] = traceLog.GetCurrentSpanId()
	header[HTTP_KEY_UNIQ_ID] = traceLog.UniqId
	header[HTTP_KEY_CALLER] = Project //注册Log时服务名称
	//header[HTTP_KEY_CALLER_URI] = traceLog.Path //当前的api接口

	header[HTTP_KEY_PRODUCT] = traceLog.Product
	header[HTTP_KEY_USERIP] = traceLog.UserIP

	//设置本应用请求时设置header头，保持链路
	header[HTTP_USER_AGENT] = _USER_AGENT + Project
	return header
}
