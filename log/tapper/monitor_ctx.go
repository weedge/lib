package tapper

import (
	"context"
	"fmt"
	"sync"

	hack "github.com/weedge/lib/strings"

	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
)

const (
	MONITORCTX = "MONITORCTX"
)

type Monitor struct {
	data map[string]interface{}
	mu   sync.Mutex
}

func (monitor *Monitor) Marshal() string {
	monitor.mu.Lock()
	defer monitor.mu.Unlock()

	b, _ := jsoniter.MarshalToString(monitor.data)
	return b
}
func (monitor *Monitor) Push(args ...interface{}) *Monitor {
	monitor.mu.Lock()
	defer monitor.mu.Unlock()

	for i := 0; i < len(args)-1; i += 2 {
		switch args[i+1].(type) {
		case string, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			monitor.data[fmt.Sprintf("%v", args[i])] = args[i+1]
		default:
			if b, e := jsoniter.Marshal(args[i+1]); e == nil {
				monitor.data[fmt.Sprintf("%v", args[i])] = hack.String(b)
			}
		}
	}
	return monitor
}

func GetMonitorFromContext(ctx context.Context) (context.Context, *Monitor) {
	switch c := ctx.(type) {
	case *gin.Context:
		if monitor, ok := c.Get(MONITORCTX); ok && monitor != nil {
			if monitor, ok := monitor.(*Monitor); ok && monitor != nil {
				return ctx, monitor
			}
		}

		monitor := &Monitor{data: map[string]interface{}{}}
		c.Set(MONITORCTX, monitor)
		return ctx, monitor

	case nil:
	default:
	}

	monitor := &Monitor{data: map[string]interface{}{}}
	return ctx, monitor
}

// 参数args: [key value]...
func PushMonitor(ctx context.Context, args ...interface{}) {
	ctx, monitor := GetMonitorFromContext(ctx)
	monitor.Push(args...)
}
