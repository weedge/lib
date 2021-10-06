package tapper

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var EmptyTraceLog = &TraceLog{}

// 生成logId
func GenLogId() string {
	now := time.Now()
	logId := ((now.UnixNano()*1e5 + now.UnixNano()/1e7) & 0x7FFFFFFF) | 0x80000000
	return strconv.FormatInt(logId, 10)
}

// 从context中获取span信息集合
func GetTraceLogFromContext(ctx context.Context) (context.Context, *TraceLog) {
	if span := ctx.Value(TRACECTX); span != nil {
		traceLog, _ := span.(*TraceLog)
		return ctx, traceLog
	}

	logId := GenLogId()
	traceLog := &TraceLog{
		LogId:  logId,
		UniqId: "",
		Caller: "",
		Refer:  "",
	}
	ctx = context.WithValue(ctx, TRACECTX, traceLog)
	return ctx, traceLog
}

// 从context中获取span信息集合
func GetTraceLogFromGinContext(ctx context.Context) (*TraceLog, bool) {
	var ok bool
	var c *gin.Context

	if c, ok = ctx.(*gin.Context); !ok {
		return EmptyTraceLog, false
	}

	var traceLog *TraceLog
	if value, exists := c.Get(TRACECTX); exists && value != nil {
		if traceLog, ok = value.(*TraceLog); traceLog != nil && ok {
			return traceLog, true
		}
	}

	return SetTraceLogFromGinHeader(c), true
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
	//设置nmq的请求transid
	traceLog.NmqTransId = c.Query("transid")

	//全程透传
	traceLog.UserIP = c.GetHeader(HTTP_KEY_USERIP)
	traceLog.Product = c.GetHeader(HTTP_KEY_PRODUCT)
	c.Set(TRACECTX, traceLog)
	return traceLog
}
