package set

import (
	"fmt"
	"math/rand"
	"testing"
)

func TestNew(t *testing.T) {
	bitSet := NewBitSet(64)
	fmt.Println(bitSet)
}

func TestString(t *testing.T) {
	bitSet := NewBitSet(256)
	bitSet.Set(64, 1)
	fmt.Println(bitSet)
}

func TestSet(t *testing.T) {
	bitSet := NewBitSet(64)
	bitSet.Set(0, 1)
	bitSet.Set(1, 1)
	bitSet.Set(63, 1)
	if bitSet.Set(64, 1) != -1 {
		t.Error("set error")
	}
	bitSet = NewBitSet(100)
	bitSet.Set(81, 1)
	if bitSet.Set(81, 1) != 1 {
		t.Error("set error")
	}
	fmt.Println(bitSet)
}

func TestSetGet(t *testing.T) {
	bitSet := NewBitSet(10000)
	bitSet.Set(1234, 1)
	if 1 != bitSet.Get(1234) {
		t.Error("set/get error")
	}
	if 0 != bitSet.Get(1235) {
		t.Error("get error")
	}
	old := bitSet.Set(1234, 0)
	if old != 1 {
		t.Error("set error")
	}
	if 0 != bitSet.Get(1234) {
		t.Error("set/get error")
	}
}

func TestBitSet_Count(t *testing.T) {
	bs := NewBitSet(5000)
	bs.Set(1, 1)
	bs.Set(101, 1)
	bs.Set(200, 1)
	bs.Set(1059, 1)
	bs.Set(3594, 1)
	bs.Set(1059, 0)
	if bs.Count() != 4 {
		t.Error("bit set count err")
	}

	//567 = 0010 0011 0111
	cn := swar(uint64(567)) //6
	if cn != 6 {
		t.Error("SWAR err")
	}
	//1234 = 0100 1101 0010
	cn = swar(uint64(1234)) //5
	if cn != 5 {
		t.Error("SWAR err")
	}
	//9874 = 0010 0110 1001 0010
	cn = swar(uint64(9874)) //6
	if cn != 6 {
		t.Error("SWAR err")
	}
	//1357913 = 0001 0100 1011 1000 0101 1001
	cn = swar(uint64(1357913)) //10
	if cn != 10 {
		t.Error("SWAR err")
	}
}

func TestBitSet_LeftShift(t *testing.T) {
	bitSet := NewBitSet(254)
	bitSet.Set(0, 1)
	fmt.Println("set 0->1", bitSet)
	bitSet.LeftShift(1)
	fmt.Println("leftShift 1", bitSet)
	bitSet.Set(63, 1)
	fmt.Println("set 63->1", bitSet)
	bitSet.Set(64, 1)
	fmt.Println("set 64->1", bitSet)
	bitSet.LeftShift(65)
	fmt.Println("leftShift 65", bitSet)
	bitSet.LeftShift(65)
	fmt.Println("leftShift 65", bitSet)
}

func TestBitSet_RightShift(t *testing.T) {
	bitSet := NewBitSet(254)
	bitSet.Set(0, 1)
	fmt.Println("set 0->1", bitSet)
	bitSet.LeftShift(1)
	fmt.Println("leftShift 1", bitSet)
	bitSet.Set(63, 1)
	fmt.Println("set 63->1", bitSet)
	bitSet.Set(64, 1)
	fmt.Println("set 64->1", bitSet)
	bitSet.LeftShift(65)
	fmt.Println("leftShift 65", bitSet)
	bitSet.RightShift(65)
	fmt.Println("rightShift 65", bitSet)
	bitSet.RightShift(65)
	fmt.Println("rightShift 65", bitSet)
}

func TestBitSet_And(t *testing.T) {
	bitSet := NewBitSet(65)
	bitSet.Set(0, 1)
	compare := NewBitSet(64)
	compare.Set(1, 1)
	res := bitSet.And(compare)
	fmt.Println("bitset:", bitSet)
	fmt.Println("compare:", compare)
	fmt.Println("bitset & compare:", res)

	println()

	bitSet = NewBitSet(64)
	bitSet.Set(0, 1)
	compare = NewBitSet(65)
	compare.Set(1, 1)
	res = bitSet.And(compare)
	fmt.Println("bitset:", bitSet)
	fmt.Println("compare:", compare)
	fmt.Println("bitset & compare:", res)

	println()

	bitSet = NewBitSet(64)
	bitSet.Set(0, 1)
	bitSet.Set(1, 1)
	compare = NewBitSet(129)
	compare.Set(1, 1)
	compare.Set(64, 1)
	res = bitSet.And(compare)
	fmt.Println("bitset:", bitSet)
	fmt.Println("compare:", compare)
	fmt.Println("bitset & compare:", res)
}

func TestBitSet_Or(t *testing.T) {
	bitSet := NewBitSet(65)
	bitSet.Set(0, 1)
	compare := NewBitSet(64)
	compare.Set(1, 1)
	res := bitSet.Or(compare)
	fmt.Println("bitset:", bitSet)
	fmt.Println("compare:", compare)
	fmt.Println("bitset | compare:", res)

	println()

	bitSet = NewBitSet(129)
	bitSet.Set(0, 1)
	compare = NewBitSet(64)
	compare.Set(1, 1)
	res = bitSet.Or(compare)
	fmt.Println("bitset:", bitSet)
	fmt.Println("compare:", compare)
	fmt.Println("bitset | compare:", res)

	println()

	bitSet = NewBitSet(65)
	bitSet.Set(0, 1)
	bitSet.Set(64, 1)
	compare = NewBitSet(129)
	compare.Set(1, 1)
	res = bitSet.Or(compare)
	fmt.Println("bitset:", bitSet)
	fmt.Println("compare:", compare)
	fmt.Println("bitset | compare:", res)
}

func TestBitSet_Xor(t *testing.T) {
	bitSet := NewBitSet(65)
	bitSet.Set(0, 1)
	compare := NewBitSet(64)
	compare.Set(1, 1)
	res := bitSet.Xor(compare)
	fmt.Println("bitset:", bitSet)
	fmt.Println("compare:", compare)
	fmt.Println("bitset ^ compare:", res)

	println()

	bitSet = NewBitSet(129)
	bitSet.Set(0, 1)
	bitSet.Set(64, 1)
	compare = NewBitSet(64)
	compare.Set(1, 1)
	res = bitSet.Or(compare)
	fmt.Println("bitset:", bitSet)
	fmt.Println("compare:", compare)
	fmt.Println("bitset ^ compare:", res)
}

func TestBitSet_Diff(t *testing.T) {
	bitSet := NewBitSet(129)
	bitSet.Set(1, 1)
	bitSet.Set(128, 1)
	compare := NewBitSet(65)
	compare.Set(1, 1)
	res := bitSet.Diff(compare)
	fmt.Println("bitset:", bitSet)
	fmt.Println("compare:", compare)
	fmt.Println("bitset ~ compare:", res)

	println()

	bitSet = NewBitSet(64)
	bitSet.Set(0, 1)
	bitSet.Set(2, 1)
	bitSet.Set(3, 1)
	fmt.Println("bitset:", bitSet)
	compare = NewBitSet(129)
	compare.Set(1, 1)
	compare.Set(2, 1)
	fmt.Println("compare:", compare)
	res = bitSet.Diff(compare)
	fmt.Println("bitset ~ compare:", res)
}

func TestBitSet_Not(t *testing.T) {
	bitSet := NewBitSet(129)
	bitSet.Set(1, 1)
	bitSet.Set(128, 1)
	fmt.Println("bitset:", bitSet)
	res := bitSet.Not()
	fmt.Println("~bitset", res)

	println()

}

func BenchmarkBitSet_Set(b *testing.B) {
	b.StopTimer()
	size := 100
	bs := NewBitSet(uint64(size))

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		pos := rand.Intn(size)
		val := rand.Intn(1)
		bs.Set(uint64(pos), val)
	}
}
