package runtimer

import (
	"fmt"
	"os"
	"runtime/debug"
	"sync"
	"time"
)

// GoSafely WaitGroup wraps a `go func()` with recover()
func GoSafely(wg *sync.WaitGroup, ignoreRecover bool, handler func(), catchFunc func(r interface{})) {
	if wg != nil {
		wg.Add(1)
	}
	go func() {
		defer func() {
			if r := recover(); r != nil {
				if !ignoreRecover {
					fmt.Fprintf(os.Stderr, "%s goroutine panic: %v\n%s\n",
						time.Now(), r, string(debug.Stack()))
				}
				if catchFunc != nil {
					if wg != nil {
						wg.Add(1)
					}
					go func() {
						defer func() {
							if p := recover(); p != nil {
								if !ignoreRecover {
									fmt.Fprintf(os.Stderr, "recover goroutine panic:%v\n%s\n",
										p, string(debug.Stack()))
								}
							}

							if wg != nil {
								wg.Done()
							}
						}()
						catchFunc(r)
					}()
				}
			}
			if wg != nil {
				wg.Done()
			}
		}()
		handler()
	}()
}

// GoUnterminated is used for which goroutine wanna long live as its process.
// @period: sleep time duration after panic to defeat @handle panic so frequently. if it is not positive,
//          the @handle will be invoked asap after panic.
func GoUnterminated(handle func(), wg *sync.WaitGroup, ignoreRecover bool, period time.Duration) {
	GoSafely(wg,
		ignoreRecover,
		handle,
		func(r interface{}) {
			if period > 0 {
				time.Sleep(period)
			}
			GoUnterminated(handle, wg, ignoreRecover, period)
		},
	)
}
