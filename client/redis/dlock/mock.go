package dlock

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"math/rand"
	"time"
)

var rdb = redis.NewClient(&redis.Options{
	Addr:     "localhost:6379",
	Password: "", // no password set
	DB:       0,  // use default DB
})

func MockTest(tag string) {
	var retryTimes = 5
	var retryInterval = time.Millisecond * 50
	lockK := "EXAMPLE_LOCK"
	expiration := time.Millisecond * 200

	locker := New(rdb, lockK, tag, retryTimes, retryInterval, expiration, true)
	locker.TryLock(context.Background())


	fmt.Println(tag + "等待业务处理完成...")
	time.Sleep(getRandDuration())

	locker.UnLock(context.Background())
}

// 生成随机时间
func getRandDuration() time.Duration {
	rand.Seed(time.Now().UnixNano())
	min := 50
	max := 100
	return time.Duration(rand.Intn(max-min)+min) * time.Millisecond
}
