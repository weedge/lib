package concurrent_map

import (
	"sync"
)

// ConcurrentMap is a thread safe map collection with better performance.
// The backend map entries are separated into the different partitions.
// Threads can access the different partitions safely without lock.
type ConcurrentMap struct {
	partitions    []*innerMap
	numOfBlockets int
}

// Partitionable is the interface which should be implemented by key type.
// It is to define how to partition the entries.
type Partitionable interface {
	// Value is raw value of the key
	Value() interface{}

	// PartitionKey is used for getting the partition to store the entry with the key.
	// E.g. the key's hash could be used as its PartitionKey
	// The partition for the key is partitions[(PartitionKey % m.numOfBlockets)]
	//
	// 1 Why not provide the default hash function for partition?
	// Ans: As you known, the partition solution would impact the performance significantly.
	// The proper partition solution balances the access to the different partitions and
	// avoid of the hot partition. The access mode highly relates to your business.
	// So, the better partition solution would just be designed according to your business.
	PartitionKey() int64
}

type innerMap struct {
	m    map[interface{}]interface{}
	lock sync.RWMutex
}

func createInnerMap() *innerMap {
	return &innerMap{
		m: make(map[interface{}]interface{}),
	}
}

func (im *innerMap) get(key Partitionable) (interface{}, bool) {
	keyVal := key.Value()
	im.lock.RLock()
	v, ok := im.m[keyVal]
	im.lock.RUnlock()
	return v, ok
}

func (im *innerMap) set(key Partitionable, v interface{}) {
	keyVal := key.Value()
	im.lock.Lock()
	im.m[keyVal] = v
	im.lock.Unlock()
}

func (im *innerMap) del(key Partitionable) {
	keyVal := key.Value()
	im.lock.Lock()
	delete(im.m, keyVal)
	im.lock.Unlock()
}

// CreateConcurrentMap is to create a ConcurrentMap with the setting number of the partitions
func CreateConcurrentMap(numOfPartitions int) *ConcurrentMap {
	var partitions []*innerMap
	for i := 0; i < numOfPartitions; i++ {
		partitions = append(partitions, createInnerMap())
	}
	return &ConcurrentMap{partitions, numOfPartitions}
}

func (m *ConcurrentMap) getPartition(key Partitionable) *innerMap {
	partitionID := key.PartitionKey() % int64(m.numOfBlockets)
	return m.partitions[partitionID]
}

// Get is to get the value by the key
func (m *ConcurrentMap) Get(key Partitionable) (interface{}, bool) {
	return m.getPartition(key).get(key)
}

// Set is to store the KV entry to the map
func (m *ConcurrentMap) Set(key Partitionable, v interface{}) {
	im := m.getPartition(key)
	im.set(key, v)
}

// Del is to delete the entries by the key
func (m *ConcurrentMap) Del(key Partitionable) {
	im := m.getPartition(key)
	im.del(key)
}

type Tuple struct {
	Key interface{}
	Val interface{}
}

// snapshot shard map fan out into channels;
// then fan in out channels for read;
func (m *ConcurrentMap) IterBuffFromSnapshot() <-chan Tuple {
	snapshotChs := m.Snapshot()
	outChCap := 0
	for _, ch := range snapshotChs {
		outChCap += cap(ch)
	}
	out := make(chan Tuple, outChCap)
	go m.fanIn(snapshotChs, out)

	return out
}

// clear all shard map
func (m *ConcurrentMap) Clear() {
	snapshotChs := m.Snapshot()
	wg := &sync.WaitGroup{}
	wg.Add(len(snapshotChs))
	for index, ch := range snapshotChs {
		go func(i int, ch chan Tuple) {
			for item := range ch {
				m.partitions[i].lock.Lock()
				delete(m.partitions[i].m, item.Key)
				m.partitions[i].lock.Unlock()
			}
			wg.Done()
		}(index, ch)
	}
	wg.Wait()
}

// snapshot shard map fan out into channels
func (m *ConcurrentMap) Snapshot() (snapshotChs []chan Tuple) {
	snapshotChs = make([]chan Tuple, m.numOfBlockets)
	wg := &sync.WaitGroup{}
	wg.Add(m.numOfBlockets)
	for i := 0; i < m.numOfBlockets; i++ {
		go func(index int, imap *innerMap) {
			imap.lock.RLock()
			snapshotChs[index] = make(chan Tuple, len(imap.m))
			for key, val := range imap.m {
				snapshotChs[index] <- Tuple{Key: key, Val: val}
			}
			imap.lock.RUnlock()
			close(snapshotChs[index]) //once write full, close then read from ch is ok
			wg.Done()
		}(i, m.partitions[i])
	}
	wg.Wait()

	return
}

func (m *ConcurrentMap) fanIn(chs []chan Tuple, out chan Tuple) {
	wg := &sync.WaitGroup{}
	wg.Add(len(chs))
	for _, ch := range chs {
		go func(ch chan Tuple) {
			for item := range ch {
				out <- item
			}
			wg.Done()
		}(ch)
	}
	wg.Wait()
	close(out) //once write full, close then read from ch is ok
}

// Count returns the number of elements within the map.
func (m ConcurrentMap) Count() int {
	count := 0
	for i := 0; i < m.numOfBlockets; i++ {
		shard := m.partitions[i]
		shard.lock.RLock()
		count += len(shard.m)
		shard.lock.RUnlock()
	}
	return count
}
