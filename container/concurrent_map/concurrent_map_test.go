package concurrent_map

import (
	"math/rand"
	"strconv"
	"testing"
)

func TestStringKeyOPs(t *testing.T) {
	for i := 0; i < 10; i++ {
		testV := rand.Intn(1000)
		m := CreateConcurrentMap(99)
		v, ok := m.Get(StrKey("Hello"))
		if v != nil || ok != false {
			t.Error("init/get failed")
		}
		m.Set(StrKey("Hello"), testV)
		v, ok = m.Get(StrKey("Hello"))
		if v.(int) != testV || ok != true {
			t.Error("set/get failed.")
		}
		m.Del(StrKey("Hello"))
		v, ok = m.Get(StrKey("Hello"))
		if v != nil || ok != false {
			t.Error("del failed")
		}

	}
}

func TestInt64KeyBasicOPs(t *testing.T) {
	for i := 0; i < 10; i++ {
		testV := rand.Int63n(1024)
		cm := CreateConcurrentMap(99)
		var key int64 = 1023
		v, ok := cm.Get(I64Key(key))
		if v != nil || ok != false {
			t.Error("init/get failed")
		}
		cm.Set(I64Key(key), testV)
		v, ok = cm.Get(I64Key(key))
		if v.(int64) != testV || ok != true {
			t.Error("set/get failed.")
		}
		cm.Del(I64Key(key))
		v, ok = cm.Get(I64Key(key))
		if v != nil || ok != false {
			t.Error("del failed")
		}
	}
}

func TestConcurrentMap_IterBuffFromSnapshot(t *testing.T) {
	m := CreateConcurrentMap(99)

	cn := 100
	for i := 0; i < cn; i++ {
		m.Set(StrKey("Hello"+strconv.Itoa(i)), i)
	}

	counter := 0
	for item := range m.IterBuffFromSnapshot() {
		val := item.Val
		if val == nil {
			t.Error("Expecting an object.")
		}
		//t.Log("iter val", val)
		counter++
	}

	if counter != cn {
		t.Error("IterBuff fail counter != cn", counter, cn)
	}
}

func GetCurrentMap(cn int) (m *ConcurrentMap) {
	m = CreateConcurrentMap(99)

	for i := 0; i < cn; i++ {
		m.Set(StrKey("Hello"+strconv.Itoa(i)), i)
	}

	return
}

func TestConcurrentMap_Clear(t *testing.T) {
	m := GetCurrentMap(100)
	cn := m.Count()
	if cn != 100 {
		t.Error("count err cn!=100")
	}

	m.Clear()

	cn = m.Count()
	if cn != 0 {
		t.Error("after clear count err cn!=0")
	}
}
