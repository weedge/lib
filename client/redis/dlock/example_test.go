package dlock

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

var rdb = redis.NewClient(&redis.Options{
	Addr:     "localhost:6379",
	Password: "", // no password set
	DB:       0,  // use default DB
})

func MockTestTryLock(tag string) {
	var retryTimes = 5
	var retryInterval = time.Millisecond * 50
	lockK := "EXAMPLE_TRY_LOCK"
	expiration := time.Millisecond * 200

	locker := New(rdb, lockK, tag, retryTimes, retryInterval, expiration, true)
	err, isGainLock := locker.TryLock(context.Background())
	if err != nil || isGainLock == false {
		return
	}

	defer locker.UnLock(context.Background())

	println(tag + " run...")
	time.Sleep(getRandDuration())
	println(tag + " over")
}

func MockTestLock(tag string) {
	var retryTimes = 10
	var retryInterval = time.Millisecond * 50
	lockK := "EXAMPLE_LOCK"
	expiration := time.Millisecond * 200

	locker := New(rdb, lockK, tag, retryTimes, retryInterval, expiration, true)
	err, isGainLock := locker.Lock(context.Background())
	if err != nil || isGainLock == false {
		return
	}

	defer locker.UnLock(context.Background())

	println(tag + " run...")
	mTime := getRandDuration()
	time.Sleep(mTime)
	println(tag + " run time " + fmt.Sprintf("%d", mTime) + " ns")
	println(tag + " over")
}

func getRandDuration() time.Duration {
	rand.Seed(time.Now().UnixNano())
	min := 50
	max := 100
	return time.Duration(rand.Intn(max-min)+min) * time.Millisecond
}

func ExampleRedisLocker_TryLock() {
	wg := &sync.WaitGroup{}
	tags := []string{"A", "B", "C", "D", "E"}
	wg.Add(len(tags))
	for _, tag := range tags {
		go func(tag string) {
			MockTestTryLock(tag)
			wg.Done()
		}(tag)
	}
	wg.Wait()

	//output:
	//
}

func ExampleRedisLocker_Lock() {
	wg := &sync.WaitGroup{}
	tags := []string{"A", "B", "C", "D", "E"}
	wg.Add(len(tags))
	for _, tag := range tags {
		go func(tag string) {
			MockTestLock(tag)
			wg.Done()
		}(tag)
	}
	wg.Wait()

	//output:
	//
}
