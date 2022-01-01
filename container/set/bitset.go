package set

import (
	"fmt"
	"math"
	"math/bits"
)

const (
	shift = 6    // 2^6 = 64
	mask  = 0x3f // 63
)

/*
type IBitSet interface {
	Set(pos int, value int) int
	Get(pos int) int
	Count() uint64
	RightShift(n int)
	LeftShift(n int)
}
*/

type BitSet struct {
	data   []uint64 //64位
	upCeil uint64   //for left/right shift
	len    uint64
	size   int
}

// 创建BitSet
func NewBitSet(len uint64) *BitSet {
	size := int(len >> shift)
	if len&mask > 0 {
		size += 1
	}
	bt := &BitSet{
		data: make([]uint64, size),
		len:  len,
		size: size,
	}
	firstSize := int(len & mask)
	for i := 0; i < firstSize; i++ {
		bt.upCeil |= 1 << i
	}
	if firstSize == 0 {
		bt.upCeil = math.MaxUint64
	}

	return bt
}

func (set *BitSet) String() string {
	//notice don't fmt print bitset obj
	return set.StringAsc()
}

func (set *BitSet) StringDesc() string {
	str := ""
	for i := set.size - 1; i >= 0; i-- {
		if i == set.size-1 {
			str += "data:"
			str += fmt.Sprintf("%b", set.data[i])
		} else {
			str += fmt.Sprintf("#%b", set.data[i])
		}
	}
	str += fmt.Sprintf(" size:%d len:%d upCeilOnesCn:%d", set.size, set.len, bits.OnesCount64(set.upCeil))

	return str
}

func (set *BitSet) StringAsc() string {
	str := ""
	for i := 0; i < set.size; i++ {
		if i == 0 {
			str += "data:"
			str += fmt.Sprintf("%b", set.data[i])
		} else {
			str += fmt.Sprintf("#%b", set.data[i])
		}
	}
	str += fmt.Sprintf(" size:%d len:%d upCeilOnesCn:%d", set.size, set.len, bits.OnesCount64(set.upCeil))

	return str
}

func (set *BitSet) StringBit() string {
	str := ""
	for i := 0; i < set.size; i++ {
		for j := mask; j >= 0; j-- {
			if set.data[i]&(uint64(1)<<j) == 1 {
				str += "1"
			} else {
				str += "0"
			}
		}
	}
	str += fmt.Sprintf(" len:%d", len(str))
	str += fmt.Sprintf("\nsize:%d  upCeilOnesCn:%d", set.size, bits.OnesCount64(set.upCeil))
	return str
}

// set in LittleEndian order
// notice: 0<= pos < len
func (set *BitSet) Set(pos uint64, value int) int {
	if pos < 0 || pos >= set.len || !(value == 0 || value == 1) {
		return -1
	}
	index, offset := set._getPos(pos)
	oldVal := set._get(index, offset)

	if value == 1 {
		set.data[index] |= uint64(1) << offset
	} else {
		set.data[index] |= uint64(1) << offset
		set.data[index] ^= uint64(1) << offset
	}

	return oldVal
}

// get in LittleEndian order (test)
// notice: 0<= pos < len
func (set *BitSet) Get(pos uint64) int {
	if pos < 0 || pos >= set.len {
		return -1
	}

	index, offset := set._getPos(pos)
	return set._get(index, offset)
}

// get data index and offset like partition offset
func (set *BitSet) _getPos(pos uint64) (index, offset int) {
	index = set.size - int(pos>>shift) - 1
	offset = int(pos & mask)

	return
}

func (set *BitSet) _get(index int, offset int) int {
	if set.data[index]&(uint64(1)<<offset) > 0 {
		return 1
	}
	return 0
}

// get in BigEndian order
func (set *BitSet) _getForBigEndian(pos int) int {
	if pos < 0 {
		return -1
	}
	index := pos >> shift
	if index >= len(set.data) {
		return -1
	}
	if set.data[index]&(1<<uint(pos&mask)) == 0 {
		return 0
	}
	return 1
}

// set in BigEndian order
func (set *BitSet) _setForBigEndian(pos int, value int) int {
	if pos < 0 || !(value == 0 || value == 1) {
		return -1
	}
	index := pos >> shift
	if index >= len(set.data) { //溢出
		return -1
	}
	oldValue := set._getForBigEndian(pos)
	if oldValue == 0 && value == 1 {
		set.data[index] |= 1 << uint(pos&mask) //对应的位设置为1，直接安位或操作即可
	} else if oldValue == 1 && value == 0 {
		set.data[index] &^= 1 << uint(pos&mask) //对应的位设置为0，先按位取反，然后进行与操作
	}
	return oldValue
}

// https://en.wikipedia.org/wiki/Hamming_weight
// use variable-precision SWAR
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

// << operator
func (set *BitSet) LeftShift(n int) {
	set.leftShiftData(n)
	set.leftShiftBit(n)
}

func (set *BitSet) leftShiftData(n int) {
	index := n >> shift
	for i := 0; i+index < set.size; i++ {
		set.data[i] = set.data[i+index]
	}
	//fmt.Println(n, index, set.data)
	for i := set.size - index; i < set.size; i++ {
		set.data[i] = 0
	}
	//fmt.Println(n, index, set.data)
}

func (set *BitSet) leftShiftBit(n int) {
	v := n & mask
	tp := uint64(0)
	lstv, pos := uint64(0), uint64(mask-v+1)
	//fmt.Println(v, tp, lstv, pos)

	for i := 1; i <= v; i++ {
		tp |= uint64(1) << (mask + 1 - i)
	}

	for i := set.size - 1; i >= 0; i-- {
		tpLstv := (set.data[i] & tp) >> pos
		set.data[i] <<= v
		set.data[i] |= lstv
		lstv = tpLstv
	}
	set.data[0] &= set.upCeil
}

// >> operator
func (set *BitSet) RightShift(n int) {
	set.rightShiftData(n)
	set.rightShiftBit(n)
}

func (set *BitSet) rightShiftData(n int) {
	index := n >> shift
	for i := set.size - 1; i >= index; i-- {
		set.data[i] = set.data[i-index]
	}
	//fmt.Println(n, index, set.data)
	for i := index - 1; i >= 0; i-- {
		set.data[i] = 0
	}
	//fmt.Println(n, index, set.data)
}

func (set *BitSet) rightShiftBit(n int) {
	v := n & mask
	tp := uint64(1)<<v - 1
	lstv, pos := uint64(0), mask-v+1
	//fmt.Println(v, tp, lstv, pos)

	for i := 0; i < set.size; i++ {
		tpLstv := (set.data[i] & tp) << pos
		set.data[i] >>= v
		set.data[i] |= lstv
		lstv = tpLstv
	}
	set.data[0] &= set.upCeil
}

// & operator (set&compare -> res)
func (set *BitSet) And(compare *BitSet) (res *BitSet) {
	panicIfNull(set)
	panicIfNull(compare)

	s, c := sortByLength(set, compare)
	res = NewBitSet(c.len)
	for i, word := range s.data {
		res.data[c.size-s.size+i] = word & c.data[c.size-s.size+i]
	}

	return
}

// | operator (set|compare -> res)
func (set *BitSet) Or(compare *BitSet) (res *BitSet) {
	panicIfNull(set)
	panicIfNull(compare)

	s, c := sortByLength(set, compare)
	res = c.Clone()
	for i, word := range s.data {
		res.data[c.size-s.size+i] = word | c.data[c.size-s.size+i]
	}

	return
}

// ^ operator (set^compare -> res)
func (set *BitSet) Xor(compare *BitSet) (res *BitSet) {
	panicIfNull(set)
	panicIfNull(compare)

	s, c := sortByLength(set, compare)
	res = c.Clone()
	for i, word := range s.data {
		res.data[c.size-s.size+i] = word ^ c.data[c.size-s.size+i]
	}

	return
}

// ~ operator(golang option ^self)
func (set *BitSet) Not() (res *BitSet) {
	panicIfNull(set)

	res = set.Clone()
	for i, word := range set.data {
		res.data[i] = ^word
	}

	return
}

// diff operator (&^) return new bitset (diff(set,compare) set not compare)
func (set *BitSet) Diff(compare *BitSet) (res *BitSet) {
	panicIfNull(set)
	panicIfNull(compare)

	// clone set (in case set is bigger than compare)
	res = set.Clone()
	if set.size > compare.size {
		for i := 0; i < compare.size; i++ {
			res.data[set.size-compare.size+i] = set.data[set.size-compare.size+i] &^ compare.data[i]
		}
	} else {
		for i := 0; i < set.size; i++ {
			res.data[i] = set.data[i] &^ compare.data[compare.size-set.size+i]
		}
	}

	return
}

// self diff operator (&^) return set diff compare(set~compare)
func (set *BitSet) InPlaceDiff(compare *BitSet) {
	panicIfNull(set)
	panicIfNull(compare)

	if set.size > compare.size {
		for i := 0; i < compare.size; i++ {
			set.data[set.size-compare.size+i] = set.data[set.size-compare.size+i] &^ compare.data[i]
		}
	} else {
		for i := 0; i < set.size; i++ {
			set.data[i] = set.data[i] &^ compare.data[compare.size-set.size+i]
		}
	}

	return
}

// Clone this BitSet
func (set *BitSet) Clone() *BitSet {
	c := NewBitSet(set.len)
	if set.data != nil { // Clone should not modify current object
		copy(c.data, set.data)
	}
	return c
}

// Convenience function: return two bitsets ordered by asc
// increasing length. Note: neither can be nil
func sortByLength(a *BitSet, b *BitSet) (ap *BitSet, bp *BitSet) {
	if a.len <= b.len {
		ap, bp = a, b
	} else {
		ap, bp = b, a
	}
	return
}

// Error is used to distinguish errors (panics) generated in this package.
type Error string

func panicIfNull(b *BitSet) {
	if b == nil {
		panic(Error("BitSet must not be null"))
	}
}
