package tapper

import (
	"bytes"
	"context"
	"strconv"

	"go.uber.org/atomic"
)

const (
	TRACECTX = "TRACECTX"
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

type TraceLog struct {
	SpanNum atomic.Int64 //用于计算本地spanid自增拼接
	LogId   string       // X_bd_logid
	SpanId  string       // X_bd_spanid
	UniqId  string       // X_bd_uniqid

	UserIP  string // X_bd_userip
	Product string // X_bd_product

	Caller string // X_bd_module
	Refer  string // X_bd_caller_uri,
	Path   string // 当前请求的地址，用作请求下游时，设置成refer

	NmqTransId string //nmq推送uri里的tranid参数，如：achilles/v3/ticker/commit?cmdno=870010&topic=core&transid=1
}

func (tl *TraceLog) FormatTraceString() string {
	spidSuffixNum := tl.SpanNum.Inc()
	bf := bytes.Buffer{}
	bf.WriteString(" [logId:")
	bf.WriteString(tl.LogId)
	bf.WriteString("] [module:")
	bf.WriteString(tl.Caller)
	bf.WriteString("] [spanid:")
	spanId := ""
	if len(tl.SpanId) > 0 {
		spanId = tl.SpanId + "." + strconv.FormatInt(spidSuffixNum, 10)
	} else {
		spanId = strconv.FormatInt(spidSuffixNum, 10)
	}
	bf.WriteString(spanId)
	if len(tl.NmqTransId) > 0 {
		bf.WriteString("] [nmq_transid:")
		bf.WriteString(tl.NmqTransId)
	}
	bf.WriteString("]")
	return bf.String()
}

func (tl *TraceLog) GetCurrentSpanId() string {
	if len(tl.SpanId) > 0 {
		return tl.SpanId + "." + strconv.FormatInt(tl.SpanNum.Load(), 10)
	}
	return strconv.FormatInt(tl.SpanNum.Load(), 10)
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
