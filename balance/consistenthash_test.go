package balance

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

var nodes []string = []string{
	"192.168.144.210:6666",
	"192.168.48.103:6666",
	"192.168.48.104:6666",
	"192.168.48.105:6666",
	"192.168.48.106:6666",
	"192.168.48.108:6666",
	"192.168.48.110:6666",
	"192.168.50.85:6666",
	"192.168.50.86:6666",
	"192.168.144.210:6667",
	"192.168.48.103:6667",
	"192.168.48.104:6667",
	"192.168.48.105:6667",
	"192.168.48.106:6667",
	"192.168.48.108:6667",
	"192.168.48.110:6667",
	"192.168.50.85:6667",
	"192.168.50.86:6667",
}

func TestAddNodes(t *testing.T) {
	hashRing := NewHashRing()
	assert.True(t, hashRing.IsEmpty())
	hashRing.AddNodes(nodes...)
	hashRing.AddNodes(nodes...)
	assert.False(t, hashRing.IsEmpty())
}

func TestGetNode(t *testing.T) {
	hashRing := NewHashRing()
	hashRing.AddNodes(nodes...)
	ipMap := make(map[string]int, 16)
	for i := 0; i < 1000000; i++ {
		key := strconv.Itoa(i) + "hasxchRingas" + strconv.Itoa(i) + "kdfsa"
		hostPort := hashRing.GetNode(key)
		ipMap[hostPort] += 1
	}
	for key, val := range ipMap {
		fmt.Printf("key:%s,val:%d\n", key, val)
	}
}
