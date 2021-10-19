package dlock

import (
	"context"
	"github.com/weedge/lib/runtimer"
	"math/rand"
	"time"

	"github.com/weedge/lib/log"

	"github.com/go-redis/redis/v8"
)

type RedisLocker struct {
	rdb           *redis.Client
	key           string
	val           int
	retryTimes    int
	retryInterval time.Duration
	expiration    time.Duration
	tag           string
	cancel        context.CancelFunc
	isWatch       bool
}

// get rand value for del lock
func getRandValue() int {
	rand.Seed(time.Now().UnixNano())
	return rand.Int()
}

func New(rdb *redis.Client, key, tag string, retryTimes int, retryInterval, expiration time.Duration, isWatch bool) *RedisLocker {
	return &RedisLocker{
		rdb:           rdb,
		key:           key,
		val:           getRandValue(),
		retryTimes:    retryTimes,
		retryInterval: retryInterval,
		expiration:    expiration,
		tag:           tag,
		cancel:        nil,
		isWatch:       isWatch,
	}
}

func (m *RedisLocker) TryLock(ctx context.Context) (isGainLock bool) {
	log.Infof("%s %s try lock", m.key, m.tag)

	set, err := rdb.SetNX(ctx, m.key, m.val, expiration).Result()
	if err != nil {
		log.Errorf("err:%s", err.Error())
	}

	// retry lock
	if set == false && m.retryLock(ctx) == false {
		log.Errorf("%s %s server unavailable, try again later", m.key, m.tag)
		return
	}
	log.Infof("%s %s try lock ok!", m.key, m.tag)

	if m.isWatch {
		var watchCtx context.Context
		watchCtx, m.cancel = context.WithCancel(context.Background())
		runtimer.GoSafely(nil, false, func() {
			m.watch(watchCtx)
		}, nil, nil)
	}

	return set
}

func (m *RedisLocker) retryLock(ctx context.Context) (isGainLock bool) {
	i := 1
	for i <= retryTimes {
		log.Infof("%s %s retry lock cn %d", m.key, m.tag, i)
		set, err := rdb.SetNX(ctx, m.key, m.val, m.expiration).Result()
		if err != nil {
			log.Errorf("err:%s", err.Error())
		}

		if set == true {
			return true
		}

		time.Sleep(m.retryInterval)
		i++
	}
	return
}

// watch to lease key expiration; or ticker to lease
func (m *RedisLocker) watch(ctx context.Context) {
	log.Infof("%s %s watching", m.key, m.tag)
	for {
		select {
		case <-ctx.Done():
			log.Infof("%s %s task done, close watch", m.key, m.tag)
			return
		default:
			// lease
			rdb.PExpire(ctx, m.key, m.expiration)
			time.Sleep(m.expiration / 2)
		}
	}
}

func (m *RedisLocker) UnLock(ctx context.Context) (isDel bool) {
	lua := `
-- 如果当前值与锁值一致,删除key
if redis.call('GET', KEYS[1]) == ARGV[1] then
	return redis.call('DEL', KEYS[1])
else
	return 0
end
`
	scriptKeys := []string{m.key}

	val, err := rdb.Eval(ctx, lua, scriptKeys, m.val).Result()
	if err != nil {
		log.Errorf("rdb.Eval error:%s", err.Error())
		return
	}

	if val == int64(1) {
		m.cancel()
		log.Infof("%s %s del ok", m.key, m.tag)
		isDel = true
	}

	return
}
