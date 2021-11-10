package sortlist

import (
	"errors"
	"sync"
	"time"

	"github.com/huandu/skiplist"
)

const (
	N_INF = "-inf"
	P_INF = "+inf"
)

type SortedList struct {
	list *skiplist.SkipList
	lock sync.RWMutex
	// 创建时间 unix timestamp in second
	createTime int64
}

func NewSortedList() *SortedList {
	return new(SortedList).Init()
}

func (sl *SortedList) Init() *SortedList {
	sl.list = skiplist.New(skiplist.BytesAsc)
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
func (sl *SortedList) Front() *skiplist.Element {
	sl.lock.RLock()
	defer sl.lock.RUnlock()
	return sl.list.Front()
}

// Back returns the last element of list l or nil if the list is empty.
func (sl *SortedList) Back() *skiplist.Element {
	sl.lock.RLock()
	defer sl.lock.RUnlock()
	return sl.list.Back()
}

func (sl *SortedList) Range(start int, stop int) []*skiplist.Element {
	if sl.Len() <= 0 {
		return nil
	}
	if offset, count, err := sl.parseLimit(start, stop); err != nil {
		return nil
	} else {
		sl.lock.RLock()
		defer sl.lock.RUnlock()
		return sl.limit(offset, count)
	}
}

// offset 从头到尾的偏移量
// count 数量
func (l *SortedList) limit(offset int, count int) []*skiplist.Element {
	if offset < 0 || count < 0 {
		return nil
	}
	res := make([]*skiplist.Element, count)
	if offset < l.list.Len() {
		e := l.list.Front()
		for i := 0; i < offset && e != nil; i++ {
			e = e.Next()
		}
		if e == nil {
			return nil
		}
		for i := 0; i < count; i++ {
			res[i] = e
			e = e.Next()
			if e == nil {
				break
			}
		}
	}
	return res
}

func (sl *SortedList) RangeByScoreAsc(min string, max string) []*skiplist.Element {
	if sl.Len() <= 0 {
		return nil
	}
	if minRes, maxRes, err := sl.parseScoreLimit(min, max); err != nil {
		return nil
	} else {
		sl.lock.RLock()
		defer sl.lock.RUnlock()
		return sl.limitByKeyAsc(minRes, maxRes)
	}
}

func (sl *SortedList) limitByKeyAsc(min string, max string) []*skiplist.Element {
	res := []*skiplist.Element{}
	for e := sl.list.Front(); e != nil && string(e.Key().([]byte)) <= max; e = e.Next() {
		if string(e.Key().([]byte)) >= min {
			res = append(res, e)
		}
	}

	return res
}

// eg: from redis zset zrange get [][]byte
func (sl *SortedList) AddBatchForStringValScores(values [][]byte) error {
	if len(values)%2 != 0 {
		return errors.New("param error")
	}
	sl.lock.Lock()
	defer sl.lock.Unlock()
	for i := 0; i < len(values); i += 2 {
		// score,val
		sl.list.Set(values[i+1], values[i])
	}

	return nil
}

// 分析start及stop
// offset 从头到尾的偏移量，大于0
// count 数目，大于0
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

// eg: redis "-inf" - "+inf"
func (sl *SortedList) parseScoreLimit(min string, max string) (minRes string, maxRes string, err error) {
	if min > max {
		err = errors.New("invalid min or max")
		return
	}
	minRes = min
	maxRes = max

	frontScore := string(sl.Front().Key().([]byte))
	backScore := string(sl.Back().Key().([]byte))
	if min == N_INF {
		minRes = frontScore
	}
	if max == P_INF {
		maxRes = backScore
	}
	return
}
