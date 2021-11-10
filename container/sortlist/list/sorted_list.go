package list

import (
	"bytes"
	"errors"
	"sync"
	"time"
)

const (
	N_INF = "-inf"
	P_INF = "+inf"
)

type SortedList struct {
	list *List
	lock sync.RWMutex
	// 创建时间 unix timestamp in second
	createTime int64
}

func NewSortedList() *SortedList {
	return new(SortedList).Init()
}

func (sl *SortedList) Init() *SortedList {
	sl.list = New()
	sl.createTime = time.Now().Unix()
	return sl
}

func (sl *SortedList) Len() int {
	sl.lock.RLock()
	defer sl.lock.RUnlock()
	return sl.list.Len()
}

// 返回链表创建时间
func (sl *SortedList) CreateTime() int64 {
	sl.lock.RLock()
	defer sl.lock.RUnlock()
	return sl.createTime
}

// Front returns the first element of list l or nil if the list is empty.
func (sl *SortedList) Front() *Element {
	sl.lock.RLock()
	defer sl.lock.RUnlock()
	return sl.list.Front()
}

// Back returns the last element of list l or nil if the list is empty.
func (sl *SortedList) Back() *Element {
	sl.lock.RLock()
	defer sl.lock.RUnlock()
	return sl.list.Back()
}

func (sl *SortedList) Range(start int, stop int) []*Element {
	if sl.Len() <= 0 {
		return nil
	}
	if offset, count, err := sl.parseLimit(start, stop); err != nil {
		return nil
	} else {
		sl.lock.RLock()
		defer sl.lock.RUnlock()
		return sl.list.Range(offset, count)
	}
}

func (sl *SortedList) RangeByScore(min string, max string) []*Element {
	if sl.Len() <= 0 {
		return nil
	}
	if minRes, maxRes, err := sl.parseScoreLimit(min, max); err != nil {
		return nil
	} else {
		sl.lock.RLock()
		defer sl.lock.RUnlock()
		return sl.list.RangeByScore(minRes, maxRes)
	}
}

func (sl *SortedList) AddBatch(values [][]byte) error {
	if len(values)%2 != 0 {
		return errors.New("param error")
	}
	sl.lock.Lock()
	defer sl.lock.Unlock()
	for i := 0; i < len(values); i += 2 {
		sl.add(values[i], values[i+1])
	}
	return nil
}

func (sl *SortedList) add(value interface{}, score []byte) {

	if llen := sl.list.len; llen <= 0 {
		sl.list.PushFront(value, score)
	} else {
		first := sl.list.Front()
		if compareScore(score, first.Score) < 0 {
			// 插入为第一个节点
			sl.list.InsertBefore(value, score, first)
		} else {
			// 从后往前遍历
			location := sl.list.Back()
			for ; compareScore(score, location.Score) < 0; location = location.Prev() {
				// 什么都不做
			}
			// 已遍历一遍到头
			sl.list.InsertAfter(value, score, location)
		}
	}
}

// 分析start及stop
// return: offset 从头到尾的偏移量，大于0
// 		   count 数目，大于0
func (sl *SortedList) parseLimit(start int, stop int) (offset int, count int, err error) {
	llen := sl.Len()
	if start < 0 {
		// 转为正数的start
		if start = llen + start; start < 0 {
			start = 0
		}
	}
	if stop < 0 {
		stop = llen + stop
	} else if stop > llen-1 {
		// 若stop参数值比list最大下标还要大，将stop作为list的最大下标来处理
		stop = llen - 1
	}
	if start > llen || start > stop {
		err = errors.New("invalid start or stop")
		return
	}

	if start < 0 || stop < 0 {
		err = errors.New("invalid start or stop")
		return
	}

	offset = start
	count = (stop - start) + 1
	return
}

func (sl *SortedList) parseScoreLimit(min string, max string) (minRes string, maxRes string, err error) {
	frontScore := string(sl.Front().Score)
	backScore := string(sl.Back().Score)
	if min == N_INF {
		min = frontScore
	}
	if max == P_INF {
		max = backScore
	}
	if min < max {
		err = errors.New("invalid min or max")
		return
	}
	if min < frontScore {
		minRes = frontScore
	}
	if max > backScore {
		maxRes = backScore
	}
	return
}

// 比较 a、b 的score值
// The result will be 0 if a==b, -1 if a < b, and +1 if a > b.
func compareScore(a, b []byte) int {
	if len(a) > len(b) {
		return 1
	} else if len(a) < len(b) {
		return -1
	} else {
		return bytes.Compare(a, b)
	}
}
