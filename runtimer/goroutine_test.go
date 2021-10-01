package runtimer

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

import (
	"github.com/stretchr/testify/assert"
)

func TestGoSafe(t *testing.T) {
	times := int32(1)

	var wg sync.WaitGroup
	GoSafely(&wg,
		false,
		func() {
			panic("hello")
		},
		func(r interface{}) {
			atomic.AddInt32(&times, 1)
		}, nil,
	)

	wg.Wait()
	assert.True(t, atomic.LoadInt32(&times) == 2)

	GoSafely(nil,
		false,
		func() {
			panic("hello")
		},
		func(r interface{}) {
			atomic.AddInt32(&times, 1)
		}, nil,
	)
	time.Sleep(1e9)
	assert.True(t, atomic.LoadInt32(&times) == 3)
}

func TestGoUnterminated(t *testing.T) {
	times := uint64(1)
	var wg sync.WaitGroup
	GoUnterminated(
		func() {
			if atomic.AddUint64(&times, 1) == 2 {
				panic("hello")
			}
		},
		&wg,
		false,
		1e8, nil,
	)
	wg.Wait()
	assert.True(t, atomic.LoadUint64(&times) == 3)

	GoUnterminated(func() {
		atomic.AddUint64(&times, 1)
	},
		nil,
		false,
		1e8, nil,
	)
	time.Sleep(1e9)
	assert.True(t, atomic.LoadUint64(&times) == 4)
}
