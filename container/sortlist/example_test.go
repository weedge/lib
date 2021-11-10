package sortlist

import (
	"fmt"
)

func ExampleSortedList_AddBatchForStringValScores() {
	sl := NewSortedList()
	valScore := [][]byte{
		[]byte(`4`), []byte(`1.4`),
		[]byte(`2`), []byte(`1.2`),
		[]byte(`3`), []byte(`1.3`),
		[]byte(`1`), []byte(`1.1`),
	}
	err := sl.AddBatchForStringValScores(valScore)
	if err != nil {
		return
	}

	fmt.Println(sl.Len())
	//fmt.Println(sl.CreateTime())
	fmt.Println(string(sl.Front().Key().([]byte)))
	//fmt.Println(sl.Front().Score())
	fmt.Println(string(sl.Back().Key().([]byte)))
	//fmt.Println(sl.Back().Score())
	res := sl.Range(0, 2)
	for _, e := range res {
		fmt.Println(string(e.Key().([]byte)))
	}

	//Output:
	// 4
	// 1.1
	// 1.4
	// 1.1
	// 1.2
	// 1.3
}

func ExampleSortedList_RangeByScoreAsc() {
	sl := NewSortedList()
	valScore := [][]byte{
		[]byte(`4`), []byte(`1.4`),
		[]byte(`2`), []byte(`1.2`),
		[]byte(`3`), []byte(`1.3`),
		[]byte(`1`), []byte(`1.1`),
	}
	err := sl.AddBatchForStringValScores(valScore)
	if err != nil {
		return
	}

	res := sl.RangeByScoreAsc("1.0","1.2")
	for _, e := range res {
		fmt.Println(string(e.Key().([]byte)))
	}

	// Output:
	// 1.1
	// 1.2
}
