package set

import (
	"testing"
)

func TestGetSet(t *testing.T) {
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
