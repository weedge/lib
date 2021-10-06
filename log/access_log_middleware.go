package log

import (
	"bytes"
	"context"
	"io/ioutil"
	"strings"
	"time"

	"github.com/weedge/lib/log/tapper"
	"github.com/weedge/lib/net"
	hack "github.com/weedge/lib/strings"

	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

const (
	LOGID     = "logId"
	REFERER   = "referer"
	COOKIE    = "cookie"
	CLIENT_IP = "client_ip"
	LOCAL_IP  = "local_ip"
	MODULE    = "module"
	UA        = "ua"
	HOST      = "host"
	URI       = "uri"
	NOTICES   = "notice"
	MONITOR   = "monitor"
	RESPONSE  = "response"
	REQUEST   = "request"
	CODE      = "code"
	COST      = "cost"
	METHOD    = "method"
	ERR       = "err"
)

const (
	REQUEST_PARAM_CTX = "REQUEST_PARAM_CTX"
)

var (
	LOCALIP        = net.GetLocalIPv4()
	MaxRespLen     = 0
	MaxReqParamLen = 4096
	IgnoreReqUris  = make([]string, 0)
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) WriteString(s string) (int, error) {
	if w.body != nil {
		w.body.WriteString(s)
	}
	return w.ResponseWriter.WriteString(s)
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	if w.body != nil {
		w.body.Write(b)
	}
	return w.ResponseWriter.Write(b)
}
func SetLogPrintMaxRespLen(maxRespLen int) {
	MaxRespLen = maxRespLen
}
func SetLogPrintMaxReqParamLen(maxReqParamLen int) {
	MaxReqParamLen = maxReqParamLen
}
func SetLogRequestParam(ctx context.Context, body interface{}) {
	data := ""
	switch body := body.(type) {
	case string:
		data = body
	default:
		if b, e := jsoniter.Marshal(body); e == nil {
			data = hack.String(b)
		}
	}

	switch c := ctx.(type) {
	case *gin.Context:
		c.Set(REQUEST_PARAM_CTX, data)
	case nil:
	default:
	}
}

// add ignore request uri
func AddIgnoreReqUri(uri ...string) {
	IgnoreReqUris = append(IgnoreReqUris, uri...)
}

// access日志打印
func GinLogger() gin.HandlerFunc {
	// 本地IP
	// 当前模块名
	return func(c *gin.Context) {
		// 开始时间
		start := time.Now()
		// 请求url
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		if raw != "" {
			path = path + "?" + raw
		}
		// 请求报文
		body, _ := ioutil.ReadAll(c.Request.Body)
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))

		// 获取当前context
		// tracer可选配
		traceLog, _ := tapper.GetTraceLogFromGinContext(c)

		blw := new(bodyLogWriter)
		if MaxRespLen <= 0 {
			blw = &bodyLogWriter{body: nil, ResponseWriter: c.Writer}
		} else {
			blw = &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		}
		c.Writer = blw

		// 处理请求
		c.Next()

		// 获取响应报文
		response := ""
		if blw.body != nil {
			response = blw.body.String()
		}

		bodyStr := ""
		flag := false
		if v, ok := c.Get(REQUEST_PARAM_CTX); ok {
			switch v := v.(type) {
			case string:
				bodyStr = v
				flag = true
			}
		}
		if !flag {
			for _, val := range IgnoreReqUris {
				if strings.Contains(path, val) {
					bodyStr = ""
					flag = true
					break
				}
			}
		}
		if !flag {
			bodyStr = string(body)
		}

		refer := c.Request.Referer()
		if len(refer) <= 0 {
			refer = traceLog.Refer
		}

		// 结束时间
		end := time.Now()
		// 执行时间 单位:微秒
		latency := end.Sub(start).Nanoseconds() / 1e3

		_, notice := tapper.GetNoticeFromContext(c)
		_, monitor := tapper.GetMonitorFromContext(c)

		fields := []zap.Field{
			zap.String(LOGID, traceLog.LogId),
			zap.String(URI, path),
			zap.String(REFERER, refer),
			zap.Any(COOKIE, c.Request.Cookies()),
			zap.String(CLIENT_IP, c.ClientIP()),
			zap.String(LOCAL_IP, LOCALIP),
			zap.String(MODULE, traceLog.Caller),
			zap.String("request_param", _trancate(bodyStr, MaxReqParamLen)),
			zap.String(UA, c.Request.UserAgent()),
			zap.String(HOST, c.Request.Host),
			zap.String(NOTICES, notice.Marshal()),
			zap.String(MONITOR, monitor.Marshal()),
			zap.Int(CODE, c.Writer.Status()),
			zap.String(RESPONSE, _trancate(response, MaxRespLen)),
			zap.Int64(COST, latency),
		}

		AccessInfo("", fields...)
	}
}

// access日志打印
func GrpcLogger() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// 开始时间
		start := time.Now()
		var traceLog *tapper.TraceLog
		// 添加context信息
		ctx, traceLog = tapper.GetTraceLogFromContext(ctx)
		resp, err := handler(ctx, req)
		// 结束时间
		end := time.Now()
		// 执行时间 单位:微秒
		latency := end.Sub(start).Nanoseconds() / 1e3
		AccessInfo("",
			zap.Any(LOGID, traceLog.LogId),
			zap.String(METHOD, info.FullMethod),
			zap.Any(REQUEST, req),
			zap.String(NOTICES, ""),
			zap.Any(RESPONSE, resp),
			zap.Any(ERR, err),
			zap.Int64(COST, latency))
		return resp, err
	}
}

func _trancate(s string, l int) string {
	if len(s) > l {
		return s[:l]
	}
	return s
}
