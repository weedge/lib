package dlock

import (
	"context"
	"math/rand"
	"time"

	"github.com/weedge/lib/log"
	"github.com/weedge/lib/runtimer"

	"github.com/go-redis/redis/v8"
)

type RedisLocker struct {
	rdb           *redis.Client
	key           string
	val           int
	retryTimes    int           // for block lock
	retryInterval time.Duration // for block lock
	expiration    time.Duration
	tag           string // logic tag
	cancel        context.CancelFunc
	isWatch       bool // is open watch to lease key ttl
}

// get rand value for del lock
func getRandValue() int {
	rand.Seed(time.Now().UnixNano())
	return rand.Int()
}

// New get a RedisLocker instance
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

// TryLock unblock try lock
func (m *RedisLocker) TryLock(ctx context.Context) (err error, isGainLock bool) {
	set, err := m.rdb.SetNX(ctx, m.key, m.val, m.expiration).Result()
	if err != nil {
		log.Errorf("err:%s", err.Error())
		return
	}

	// retry lock
	if set == false {
		log.Infof("%s %s try lock fail", m.key, m.tag)
		return
	}

	log.Infof("%s %s try lock ok", m.key, m.tag)

	if m.isWatch {
		var watchCtx context.Context
		watchCtx, m.cancel = context.WithCancel(context.Background())
		runtimer.GoSafely(nil, false, func() {
			m.watch(watchCtx)
		}, nil, nil)
	}

	return nil, set
}

// Lock block Lock util retryTimes per retryInterval
func (m *RedisLocker) Lock(ctx context.Context) (err error, isGainLock bool) {
	set, err := m.rdb.SetNX(ctx, m.key, m.val, m.expiration).Result()
	if err != nil {
		log.Errorf("err:%s", err.Error())
		return
	}

	if set == false {
		err, isGainLock = m.retryLock(ctx)
		if err != nil {
			return
		}

		if isGainLock == false {
			log.Infof("%s %s lock fail", m.key, m.tag)
			return
		}
	}
	log.Infof("%s %s lock ok", m.key, m.tag)

	if m.isWatch {
		var watchCtx context.Context
		watchCtx, m.cancel = context.WithCancel(context.Background())
		runtimer.GoSafely(nil, false, func() {
			m.watch(watchCtx)
		}, nil, nil)
	}

	return nil, true
}

// retry lock util retry times by retry interval or gain lock return
func (m *RedisLocker) retryLock(ctx context.Context) (err error, isGainLock bool) {
	i := 1
	var set bool
	for {
		if i > m.retryTimes {
			break
		}
		if m.retryTimes > 0 {
			log.Infof("%s %s retry lock cn %d", m.key, m.tag, i)
			i++
		}

		set, err = m.rdb.SetNX(ctx, m.key, m.val, m.expiration).Result()
		if err != nil {
			return
		}
		if set == true {
			isGainLock = set
			return
		}

		time.Sleep(m.retryInterval)
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
			m.rdb.PExpire(ctx, m.key, m.expiration)
			time.Sleep(m.expiration / 2)
		}
	}
}

// UnLock unlock ok return true or false by lua script for atomic cmd(get->del)
func (m *RedisLocker) UnLock(ctx context.Context) (isDel bool) {
	lua := `
if redis.call('GET', KEYS[1]) == ARGV[1] then
	return redis.call('DEL', KEYS[1])
else
	return 0
end
`
	scriptKeys := []string{m.key}

	val, err := m.rdb.Eval(ctx, lua, scriptKeys, m.val).Result()
	if err != nil {
		log.Errorf("rdb.Eval error:%s", err.Error())
		return
	}

	if val == int64(1) {
		if m.cancel != nil {
			m.cancel()
		}
		log.Infof("%s %s unlock ok", m.key, m.tag)
		isDel = true
	}

	return
}
