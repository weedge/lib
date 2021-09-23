package sort_map

import (
	"fmt"
	"sort"
)

func wordCnSortByChar(longStr string) {
	mapChars := make(map[byte]int)
	chars := make([]int, 0, len(longStr))
	for _, c := range longStr {
		if _, ok := mapChars[byte(c)]; !ok {
			chars = append(chars, int(c))
		}
		mapChars[byte(c)]++
	}
	for k, v := range mapChars {
		println(k, v)
	}

	sort.Ints(chars)
	/*
		sort.Slice(chars, func(i, j int) bool {
			if chars[i] < chars[j] {
				return true
			}
			return false
		})
	*/
	println()
	for _, char := range chars {
		fmt.Println(fmt.Sprintf("%s", []byte{byte(char)}), mapChars[byte(char)])
	}
}

func wordCnSortByCharSlice(longStr string) {
	mapChars := make(map[int64]int64)
	for _, c := range longStr {
		mapChars[int64(c)]++
	}
	for k, v := range mapChars {
		println(k, v)
	}
	println()
	sliceKVChars := SortIntIntMapByKey(mapChars)
	for _, item := range sliceKVChars {
		fmt.Println(fmt.Sprintf("%s", []byte{byte(item.Key)}), item.Value)
	}
}

func ExampleSortIntIntMapByKey() {
	longStr := "aabbccddddca"
	wordCnSortByChar(longStr)
	wordCnSortByCharSlice(longStr)
	// Output:
	// a 3
	// b 2
	// c 3
	// d 4
	// a 3
	// b 2
	// c 3
	// d 4
}
