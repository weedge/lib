package dlock

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"

	"github.com/weedge/lib/log"
)

func Example_Lock() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Error(err)
	}
	defer cli.Close()

	lockKey := "/lock"

	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		session, err := concurrency.NewSession(cli)
		if err != nil {
			log.Error(err)
		}
		defer session.Close()
		m := concurrency.NewMutex(session, lockKey)
		if err := m.Lock(context.TODO()); err != nil {
			log.Error("go1 get mutex failed " + err.Error())
		}
		println(fmt.Sprintf("go1 get mutex sucess"))
		println(fmt.Sprintf("%+v", m))
		time.Sleep(time.Duration(10) * time.Second)
		m.Unlock(context.TODO())
		println("go1 release lock")
	}()

	go func() {
		defer wg.Done()
		session, err := concurrency.NewSession(cli)
		if err != nil {
			log.Error(err)
		}
		m := concurrency.NewMutex(session, lockKey)
		if err := m.Lock(context.TODO()); err != nil {
			log.Error("go2 get mutex failed " + err.Error())
		}
		println("go2 get mutex success")
		println(fmt.Sprintf("%+v", m))
		time.Sleep(time.Duration(2) * time.Second)
		m.Unlock(context.TODO())
		println("go2 release lock")
	}()

	wg.Wait()

	println("over")
	// Output:
	//
}

func Example_TryLock() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Error(err)
	}
	defer cli.Close()

	lockKey := "/lock"

	wg := &sync.WaitGroup{}
	wg.Add(2)

	// tryLock
	go func() {
		defer wg.Done()
		session, err := concurrency.NewSession(cli)
		if err != nil {
			log.Error(err)
		}
		defer session.Close()
		m := concurrency.NewMutex(session, lockKey)
		err1 := m.TryLock(context.TODO())
		if err1 != nil {
			println("cannot acquire lock for go3, as already locked in another session", err1.Error())
		} else {
			println("go3 get mutex success")
			println(fmt.Sprintf("%+v", m))
			time.Sleep(time.Duration(2) * time.Second)
			m.Unlock(context.TODO())
			println("sleep 2s go3 release lock")
		}
	}()

	// tryLock
	go func() {
		defer wg.Done()
		session, err := concurrency.NewSession(cli)
		if err != nil {
			log.Error(err)
		}
		defer session.Close()
		m := concurrency.NewMutex(session, lockKey)
		err1 := m.TryLock(context.TODO())
		if err1 != nil {
			println("cannot acquire lock for go4, as already locked in another session", err1.Error())
		} else {
			println("go4 get mutex success")
			println(fmt.Sprintf("%+v", m))
			time.Sleep(time.Duration(2) * time.Second)
			m.Unlock(context.TODO())
			println("sleep 2s go4 release lock")
		}
	}()

	wg.Wait()

	println("over")
	// Output:
	//
}
