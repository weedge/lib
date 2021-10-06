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
	NOTICECTX = "NOTICECTX"
)

type Notice struct {
	data map[string]interface{}
	mu   sync.Mutex
}

func (notice *Notice) Marshal() string {
	notice.mu.Lock()
	defer notice.mu.Unlock()

	b, _ := jsoniter.MarshalToString(notice.data)
	return b
}
func (notice *Notice) Push(args ...interface{}) *Notice {
	notice.mu.Lock()
	defer notice.mu.Unlock()

	for i := 0; i < len(args)-1; i += 2 {
		switch args[i+1].(type) {
		case string, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			notice.data[fmt.Sprintf("%v", args[i])] = args[i+1]
		default:
			if b, e := jsoniter.Marshal(args[i+1]); e == nil {
				notice.data[fmt.Sprintf("%v", args[i])] = hack.String(b)
			}
		}
	}
	return notice
}

func GetNoticeFromContext(ctx context.Context) (context.Context, *Notice) {
	switch c := ctx.(type) {
	case *gin.Context:
		if notice, ok := c.Get(NOTICECTX); ok && notice != nil {
			if notice, ok := notice.(*Notice); ok && notice != nil {
				return ctx, notice
			}
		}

		notice := &Notice{data: map[string]interface{}{}}
		c.Set(NOTICECTX, notice)
		return ctx, notice

	case nil:
	default:
	}

	notice := &Notice{data: map[string]interface{}{}}
	return ctx, notice
}

// 参数args: [key value]...
func PushNotice(ctx context.Context, args ...interface{}) {
	ctx, notice := GetNoticeFromContext(ctx)
	notice.Push(args...)
}
