package tapper

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"sync"

	hack "github.com/weedge/lib/strings"

	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"go.uber.org/atomic"
)

const (
	TRACECTX = "TRACECTX"
	TRACELOG = "TRACELOG"
)

type Trace struct {
	data map[string]interface{}
	mu   sync.Mutex
}

func (trace *Trace) Marshal() string {
	trace.mu.Lock()
	defer trace.mu.Unlock()

	b, _ := jsoniter.MarshalToString(trace.data)
	return b
}
func (trace *Trace) Push(args ...interface{}) *Trace {
	trace.mu.Lock()
	defer trace.mu.Unlock()

	for i := 0; i < len(args)-1; i += 2 {
		switch args[i+1].(type) {
		case string, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			trace.data[fmt.Sprintf("%v", args[i])] = args[i+1]
		default:
			if b, e := jsoniter.Marshal(args[i+1]); e == nil {
				trace.data[fmt.Sprintf("%v", args[i])] = hack.String(b)
			}
		}
	}
	return trace
}

func GetTraceFromContext(ctx context.Context) (context.Context, *Trace) {
	switch c := ctx.(type) {
	case *gin.Context:
		if trace, ok := c.Get(TRACECTX); ok && trace != nil {
			if trace, ok := trace.(*Trace); ok && trace != nil {
				return ctx, trace
			}
		}

		trace := &Trace{data: map[string]interface{}{}}
		c.Set(TRACECTX, trace)
		return ctx, trace

	case nil:
	default:
	}

	trace := &Trace{data: map[string]interface{}{}}
	return ctx, trace
}

// 参数args: [key value]...
func PushTrace(ctx context.Context, args ...interface{}) {
	ctx, trace := GetTraceFromContext(ctx)
	trace.Push(args...)
}

type TraceLog struct {
	SpanNum atomic.Int64 //用于计算本地spanId自增拼接
	LogId   string
	SpanId  string
	UniqId  string

	UserIP  string
	Product string

	Caller string
	Refer  string
	Path   string // 当前请求的地址，用作请求下游时，设置成refer

	MqTransId string //mq tranId 比如mq push模式 push uri地址参数中会有带上事务id
}

func (tl *TraceLog) FormatTraceString() string {
	spidSuffixNum := tl.SpanNum.Inc()
	bf := bytes.Buffer{}
	bf.WriteString(" [logId:")
	bf.WriteString(tl.LogId)
	bf.WriteString("] [module:")
	bf.WriteString(tl.Caller)
	bf.WriteString("] [spanId:")
	spanId := ""
	if len(tl.SpanId) > 0 {
		spanId = tl.SpanId + "." + strconv.FormatInt(spidSuffixNum, 10)
	} else {
		spanId = strconv.FormatInt(spidSuffixNum, 10)
	}
	bf.WriteString(spanId)
	if len(tl.MqTransId) > 0 {
		bf.WriteString("] [mq_transId:")
		bf.WriteString(tl.MqTransId)
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

var EmptyTraceLog = &TraceLog{}

// 从context中获取span信息集合
func GetTraceLogFromContext(ctx context.Context) (context.Context, *TraceLog) {
	if span := ctx.Value(TRACELOG); span != nil {
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
	ctx = context.WithValue(ctx, TRACELOG, traceLog)
	return ctx, traceLog
}

// 从gin context中获取span信息集合
func GetTraceLogFromGinContext(ctx context.Context) (*TraceLog, bool) {
	var ok bool
	var c *gin.Context

	if c, ok = ctx.(*gin.Context); !ok {
		return EmptyTraceLog, false
	}

	var traceLog *TraceLog
	if value, exists := c.Get(TRACELOG); exists && value != nil {
		if traceLog, ok = value.(*TraceLog); traceLog != nil && ok {
			return traceLog, true
		}
	}

	return TraceLogger.SetTraceLogFromGinHeader(c), true
}

