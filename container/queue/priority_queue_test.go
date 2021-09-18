package queue

import (
	"container/heap"
	"math/rand"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"testing"
)

func equal(t *testing.T, act, exp interface{}) {
	if !reflect.DeepEqual(exp, act) {
		_, file, line, _ := runtime.Caller(1)
		t.Logf("\033[31m%s:%d:\n\n\texp: %#v\n\n\tgot: %#v\033[39m\n\n",
			filepath.Base(file), line, exp, act)
		t.FailNow()
	}
}

func TestPriorityQueue_Ops(t *testing.T) {
	c := 100
	pq := NewPriorityQueue(c)

	ints := make([]int, 0, c)
	for i := 0; i < c; i++ {
		v := rand.Int()
		ints = append(ints, v)
		heap.Push(&pq, &Item{Value: "test", Priority: int64(v)})
	}
	equal(t, len(pq), c)

	sort.Ints(ints)

	item := heap.Pop(&pq)
	equal(t, item.(*Item).Priority, int64(ints[0]))

	for i := 0; i < 10; i++ {
		heap.Remove(&pq, rand.Intn((c-1)-i))
	}

	lastPriority := heap.Pop(&pq).(*Item).Priority
	for i := 0; i < (c - 10 - 1 - 1); i++ {
		item := heap.Pop(&pq)
		equal(t, lastPriority < item.(*Item).Priority, true)
		lastPriority = item.(*Item).Priority
	}

	equal(t, len(pq), 0)
}

func TestPriorityQueue_PeekAndShift(t *testing.T) {
	c := 100
	pq := NewPriorityQueue(c)

	for i := 0; i < c; i++ {
		heap.Push(&pq, &Item{Value: "test", Priority: int64(i)})
	}
	equal(t, len(pq), c)

	for i := 0; i < c; i++ {
		item, _ := pq.PeekAndShift(int64(c - 1))
		equal(t, item.Priority, int64(i))
	}
}

func TestPriorityQueue_Update(t *testing.T) {
	c := 10
	pq := NewPriorityQueue(c)

	for i := 0; i < c; i++ {
		heap.Push(&pq, &Item{Value: "test", Priority: int64(i)})
	}
	equal(t, len(pq), c)
	equal(t, pq.Top().Priority, int64(0))

	item := &Item{Value: "test-1", Priority: int64(-1)}
	heap.Push(&pq, item)
	equal(t, pq.Top().Priority, int64(-1))

	pq.Update(item, "test10", int64(c))
	equal(t, pq.Top().Priority, int64(0))

	equal(t, len(pq), c+1)
}
