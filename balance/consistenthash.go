package balance

// 一致性HASH负载，采用CRC32算法，
// 单调性:环形空间，0 ~ 2^32-1次方的数值空间
// 平衡性: 哈希算法并不能保证100%的平衡性,引入虚拟节点分散负载
import (
	"fmt"
	"hash/crc32"
	"sort"
	"strconv"
	"sync"
)

type UInt32Slice []uint32

func (s UInt32Slice) Len() int {
	return len(s)
}

func (s UInt32Slice) Less(i, j int) bool {
	return s[i] < s[j]
}

func (s UInt32Slice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type Hash func(data []byte) uint32

type HashRing struct {
	hash     Hash
	replicas int               // 复制因子,虚拟节点
	keys     UInt32Slice       // 已排序的节点哈希切片
	Nodes    map[uint32]string // 节点哈希和KEY的HashRing，键是哈希值，值是节点Key
	mutex    sync.RWMutex
}

func NewHashRing() *HashRing {
	hashRing := &HashRing{
		replicas: 400,                // 复制因子,虚拟节点200个，改动此处会影响已有key的hash的节点
		hash:     crc32.ChecksumIEEE, //使用CRC32算法
		Nodes:    make(map[uint32]string),
	}
	return hashRing
}

func (hashRing *HashRing) IsEmpty() bool {
	return len(hashRing.keys) == 0
}

// Add 方法用来添加缓存节点，参数为节点key，比如使用IP
func (hashRing *HashRing) AddNodes(nodes ...string) {
	hashRing.mutex.Lock()
	defer hashRing.mutex.Unlock()
	for _, node := range nodes {
		// 结合复制因子计算所有虚拟节点的hash值，并存入m.keys中，同时在m.Nodes中保存哈希值和key的映射
		for i := 0; i < hashRing.replicas; i++ {
			hash := hashRing.hash([]byte(node + strconv.Itoa(i)))
			if _, ok := hashRing.Nodes[hash]; !ok {
				hashRing.keys = append(hashRing.keys, hash)
				hashRing.Nodes[hash] = node
			}
		}
	}
	// 对所有虚拟节点的哈希值进行排序，方便之后进行二分查找
	sort.Sort(hashRing.keys)
}

// Get 方法根据给定的对象获取最靠近它的那个节点key
func (hashRing *HashRing) GetNode(key string) string {
	if hashRing.IsEmpty() {
		return ""
	}
	hashRing.mutex.RLock()
	defer hashRing.mutex.RUnlock()
	hash := hashRing.hash([]byte(key))
	// 通过二分查找获取最优节点，第一个节点hash值大于对象hash值的就是最优节点
	idx := sort.Search(len(hashRing.keys), func(i int) bool { return hashRing.keys[i] >= hash })

	// 如果查找结果大于节点哈希数组的最大索引，表示此时该对象哈希值位于最后一个节点之后，那么放入第一个节点中
	if idx == len(hashRing.keys) {
		idx = 0
	}
	return hashRing.Nodes[hashRing.keys[idx]]
}

func (hashRing *HashRing) printKeys() {
	fmt.Println("keys:", hashRing.keys)
}
