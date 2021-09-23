package sort_map

import (
	"sort"
)

type IntIntKVPair struct {
	Key   int64
	Value int64
}
type IntStringKVPair struct {
	Key   int64
	Value string
}
type StringStringKVPair struct {
	Key   string
	Value string
}
type StringIntKVPair struct {
	Key   string
	Value int64
}

type IntIntKVPairList []IntIntKVPair
type IntStringKVPairList []IntStringKVPair
type StringStringKVPairList []StringStringKVPair
type StringIntKVPairList []StringIntKVPair

type KVPairSortList struct {
	IntIntKVPairList
	IntStringKVPairList
	StringStringKVPairList
	StringIntKVPairList

	KVType string // "int|int", "string|string", "int|string", "string|int"
	SortBy string // "key", "val", "key asc", "val asc", "key desc", "val desc",
}

func (p KVPairSortList) Swap(i, j int) {
	switch p.KVType {
	case "int|int":
		p.IntIntKVPairList[i], p.IntIntKVPairList[j] = p.IntIntKVPairList[j], p.IntIntKVPairList[i]
	case "int|string":
		p.IntStringKVPairList[i], p.IntStringKVPairList[j] = p.IntStringKVPairList[j], p.IntStringKVPairList[i]
	case "string|string":
		p.StringStringKVPairList[i], p.StringStringKVPairList[j] = p.StringStringKVPairList[j], p.StringStringKVPairList[i]
	case "string|int":
		p.StringIntKVPairList[i], p.StringIntKVPairList[j] = p.StringIntKVPairList[j], p.StringIntKVPairList[i]
	}
}

func (p KVPairSortList) Len() int {
	switch p.KVType {
	case "int|int":
		return len(p.IntIntKVPairList)
	case "int|string":
		return len(p.IntStringKVPairList)
	case "string|string":
		return len(p.StringStringKVPairList)
	case "string|int":
		return len(p.StringIntKVPairList)
	}
	return 0
}

func (p KVPairSortList) Less(i, j int) bool {
	switch p.KVType {
	case "int|int":
		return p.intIntVKLess(i, j)
	case "int|string":
		return p.intStringVKLess(i, j)
	case "string|string":
		return p.stringStringVKLess(i, j)
	case "string|int":
		return p.stringIntVKLess(i, j)
	}

	return false
}

func (p KVPairSortList) intIntVKLess(i, j int) bool {
	if p.SortBy == "val" || p.SortBy == "val asc" {
		return p.IntIntKVPairList[i].Value < p.IntIntKVPairList[j].Value
	}
	if p.SortBy == "val desc" {
		return p.IntIntKVPairList[i].Value > p.IntIntKVPairList[j].Value
	}
	if p.SortBy == "key" || p.SortBy == "key asc" {
		return p.IntIntKVPairList[i].Key < p.IntIntKVPairList[j].Key
	}
	if p.SortBy == "key desc" {
		return p.IntIntKVPairList[i].Key > p.IntIntKVPairList[j].Key
	}

	return false
}

func (p KVPairSortList) intStringVKLess(i, j int) bool {
	if p.SortBy == "val" || p.SortBy == "val asc" {
		return p.IntStringKVPairList[i].Value < p.IntStringKVPairList[j].Value
	}
	if p.SortBy == "val desc" {
		return p.IntStringKVPairList[i].Value > p.IntStringKVPairList[j].Value
	}
	if p.SortBy == "key" || p.SortBy == "key asc" {
		return p.IntStringKVPairList[i].Key < p.IntStringKVPairList[j].Key
	}
	if p.SortBy == "key desc" {
		return p.IntStringKVPairList[i].Key > p.IntStringKVPairList[j].Key
	}

	return false
}

func (p KVPairSortList) stringStringVKLess(i, j int) bool {
	if p.SortBy == "val" || p.SortBy == "val asc" {
		return p.StringStringKVPairList[i].Value < p.StringStringKVPairList[j].Value
	}
	if p.SortBy == "val desc" {
		return p.StringStringKVPairList[i].Value > p.StringStringKVPairList[j].Value
	}
	if p.SortBy == "key" || p.SortBy == "key asc" {
		return p.StringStringKVPairList[i].Key < p.StringStringKVPairList[j].Key
	}
	if p.SortBy == "key desc" {
		return p.StringStringKVPairList[i].Key > p.StringStringKVPairList[j].Key
	}

	return false
}

func (p KVPairSortList) stringIntVKLess(i, j int) bool {
	if p.SortBy == "val" || p.SortBy == "val asc" {
		return p.StringIntKVPairList[i].Value < p.StringIntKVPairList[j].Value
	}
	if p.SortBy == "val desc" {
		return p.StringIntKVPairList[i].Value > p.StringIntKVPairList[j].Value
	}
	if p.SortBy == "key" || p.SortBy == "key asc" {
		return p.StringIntKVPairList[i].Key < p.StringIntKVPairList[j].Key
	}
	if p.SortBy == "key desc" {
		return p.StringIntKVPairList[i].Key > p.StringIntKVPairList[j].Key
	}

	return false
}

// map[int64]int64 value asc
func SortIntIntMapByValue(m map[int64]int64) IntIntKVPairList {
	return sortIntIntMap(m, "val")
}

// map[int64]int64 value desc
func SortIntIntMapByValueDesc(m map[int64]int64) IntIntKVPairList {
	return sortIntIntMap(m, "val desc")
}

// map[int64]int64 key
func SortIntIntMapByKey(m map[int64]int64) IntIntKVPairList {
	return sortIntIntMap(m, "key")
}

// map[int64]int64 key desc
func SortIntIntMapByKeyDesc(m map[int64]int64) IntIntKVPairList {
	return sortIntIntMap(m, "key desc")
}

func sortIntIntMap(m map[int64]int64, sortBy string) IntIntKVPairList {
	kvPairSortList := KVPairSortList{}
	kvPairSortList.KVType = "int|int"
	kvPairSortList.SortBy = sortBy
	kvPairSortList.IntIntKVPairList = make(IntIntKVPairList, len(m))

	i := 0
	for k, v := range m {
		kvPairSortList.IntIntKVPairList[i] = IntIntKVPair{k, v}
		i++
	}

	sort.Sort(kvPairSortList)

	return kvPairSortList.IntIntKVPairList
}

// map[int64]string value asc
func SortIntStringMapByValue(m map[int64]string) IntStringKVPairList {
	return sortIntStringMap(m, "val")
}

// map[int64]string value Desc
func SortIntStringMapByValueDesc(m map[int64]string) IntStringKVPairList {
	return sortIntStringMap(m, "val desc")
}

// map[int64]string key asc
func SortIntStringMapByKey(m map[int64]string) IntStringKVPairList {
	return sortIntStringMap(m, "key")
}

// map[int64]string key Desc
func SortIntStringMapByKeyDesc(m map[int64]string) IntStringKVPairList {
	return sortIntStringMap(m, "key desc")
}

func sortIntStringMap(m map[int64]string, sortBy string) IntStringKVPairList {
	kvPairSortList := &KVPairSortList{}
	kvPairSortList.KVType = "string|int"
	kvPairSortList.SortBy = sortBy
	kvPairSortList.IntStringKVPairList = make(IntStringKVPairList, len(m))

	i := 0
	for k, v := range m {
		kvPairSortList.IntStringKVPairList[i] = IntStringKVPair{k, v}
		i++
	}

	sort.Sort(kvPairSortList)

	return kvPairSortList.IntStringKVPairList
}

// map[string]string value Desc
func SortStringStringMapByValue(m map[string]string) StringStringKVPairList {
	return sortStringStringMap(m, "val")
}

// map[string]string value Desc
func SortStringStringMapByValueDesc(m map[string]string) StringStringKVPairList {
	return sortStringStringMap(m, "val desc")
}

// map[string]string key asc
func SortStringStringMapByKey(m map[string]string) StringStringKVPairList {
	return sortStringStringMap(m, "key")
}

// map[string]string key Desc
func SortStringStringMapByKeyDesc(m map[string]string) StringStringKVPairList {
	return sortStringStringMap(m, "key desc")
}

func sortStringStringMap(m map[string]string, sortBy string) StringStringKVPairList {
	kvPairSortList := &KVPairSortList{}
	kvPairSortList.KVType = "string|string"
	kvPairSortList.SortBy = sortBy
	kvPairSortList.StringStringKVPairList = make(StringStringKVPairList, len(m))

	i := 0
	for k, v := range m {
		kvPairSortList.StringStringKVPairList[i] = StringStringKVPair{k, v}
		i++
	}

	sort.Sort(kvPairSortList)

	return kvPairSortList.StringStringKVPairList
}

// map[string]int64 value
func SortStringIntMapByValue(m map[string]int64) StringIntKVPairList {
	return sortStringIntMap(m, "val desc")
}

// map[string]int64 value desc
func SortStringIntMapByValueDesc(m map[string]int64) StringIntKVPairList {
	return sortStringIntMap(m, "val desc")
}

// map[string]int64 key asc
func SortStringIntMapByKey(m map[string]int64) StringIntKVPairList {
	return sortStringIntMap(m, "key")
}

// map[string]int64 key desc
func SortStringIntMapByKeyDesc(m map[string]int64) StringIntKVPairList {
	return sortStringIntMap(m, "key desc")
}

func sortStringIntMap(m map[string]int64, sortBy string) StringIntKVPairList {
	kvPairSortList := &KVPairSortList{}
	kvPairSortList.KVType = "string|int"
	kvPairSortList.SortBy = sortBy
	kvPairSortList.StringIntKVPairList = make(StringIntKVPairList, len(m))

	i := 0
	for k, v := range m {
		kvPairSortList.StringIntKVPairList[i] = StringIntKVPair{k, v}
		i++
	}

	sort.Sort(kvPairSortList)

	return kvPairSortList.StringIntKVPairList
}
