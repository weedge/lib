package set

import "fmt"

const (
	shift = 6    // 2^6 = 64
	mask  = 0x3f // 63
)

type BitSet struct {
	data []uint64 //64位
	size int
}

//创建BitSet
func NewBitSet(len int) *BitSet {
	size := len>>shift + 1
	return &BitSet{
		data: make([]uint64, size),
		size: size,
	}
}

func (set *BitSet) String() string {
	str := ""
	for i := set.size - 1; i >= 0; i-- {
		str += fmt.Sprintf("%d", set.data[i])
	}

	return str
}

func (set *BitSet) Get(pos int) int {
	if pos < 0 { // pos必须是正数
		return -1
	}
	index := pos >> shift
	if index >= len(set.data) { //溢出
		return -1
	}
	if set.data[index]&(1<<uint(pos&mask)) == 0 {
		return 0
	}
	return 1
}

// 类似redis bitmap，设置后返回设置前的值
func (set *BitSet) Set(pos int, value int) int {
	if pos < 0 || !(value == 0 || value == 1) {
		return -1
	}
	index := pos >> shift
	if index >= len(set.data) { //溢出
		return -1
	}
	oldValue := set.Get(pos)
	if oldValue == 0 && value == 1 {
		set.data[index] |= 1 << uint(pos&mask) //对应的位设置为1，直接安位或操作即可
	} else if oldValue == 1 && value == 0 {
		set.data[index] &^= 1 << uint(pos&mask) //对应的位设置为0，先按位取反，然后进行与操作
	}
	return oldValue
}

// 计算汉明重量
// https://en.wikipedia.org/wiki/Hamming_weight
func (set *BitSet) Count() uint64 {
	var count uint64
	for _, b := range set.data {
		count += swar(b)
	}
	return count
}

// variable-precision SWAR
func swar(i uint64) uint64 {
	// 将相邻2位的1的数量计算出来，结果存放在这2位
	i = (i & 0x5555555555555555) + ((i >> 1) & 0x5555555555555555)
	// 将相邻4位的结果相加，结果存放在这4位
	i = (i & 0x3333333333333333) + ((i >> 2) & 0x3333333333333333)
	// 将相邻8位的结果相加，结果存放在这8位
	i = (i & 0x0F0F0F0F0F0F0F0F) + ((i >> 4) & 0x0F0F0F0F0F0F0F0F)
	// 计算整体1的数量，记录在高8位，然后通过右移运算，将结果放到低8位，得到最终结果
	i = (i * 0x0101010101010101) >> 56
	return i
}

func (set *BitSet) LeftShift(n int) {
	index := n >> shift
	for i := set.size - 1; i >= index; i-- {
		set.data[i] = set.data[i-index]
	}
	for i := index - 1; i >= 0; i++ {
		set.data[i] = 0
	}
	//todo
}

func (set *BitSet) RightShift(n int) {

}
