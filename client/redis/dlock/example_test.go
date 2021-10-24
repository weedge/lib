package dlock

import (
	"context"
	"fmt"
	"github.com/weedge/lib/strings"
	"math/rand"
	"runtime/debug"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

var rdb = redis.NewClient(&redis.Options{
	Addr:         "localhost:6379",
	Password:     "", // no password set
	DB:           0,  // use default DB
	DialTimeout:  10 * time.Second,
	ReadTimeout:  30 * time.Second,
	WriteTimeout: 30 * time.Second,
	PoolSize:     10,
	PoolTimeout:  30 * time.Second,
})

func MockTestTryLock(tag string) {
	var retryTimes = 5
	var retryInterval = time.Millisecond * 50
	lockK := "EXAMPLE_TRY_LOCK"
	expiration := time.Millisecond * 200

	locker := New(rdb, lockK, tag, &UUIDVal{}, retryTimes, retryInterval, expiration, true)
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
	var retryTimes = -1
	var retryInterval = time.Millisecond * 50
	lockK := "EXAMPLE_LOCK"
	expiration := time.Millisecond * 200

	locker := New(rdb, lockK, tag, &UUIDVal{}, retryTimes, retryInterval, expiration, true)
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

func Recover() {
	if e := recover(); e != nil {
		fmt.Printf("panic: %v, stack: %v\n", e, strings.BytesToString(debug.Stack()))
	}
}

func ExampleRedisLocker_LockLongPanic() {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer Recover()
		var retryTimes = 5
		var retryInterval = time.Millisecond * 50
		lockK := "EXAMPLE_LOCK_WATCH"
		expiration := time.Millisecond * 200
		tag := "panic"

		locker := New(rdb, lockK, tag, &UUIDVal{}, retryTimes, retryInterval, expiration, true)
		defer locker.UnLock(context.Background())
		err, isGainLock := locker.TryLock(context.Background())
		if err != nil || isGainLock == false {
			return
		}

		time.Sleep(20 * time.Second)
		panic("test panic for lock watch")
	}()
	wg.Wait()
	println("over")

	//output:
	//
}
