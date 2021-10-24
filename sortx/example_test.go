package sortx

import (
	"fmt"
	"sort"
)

func ExampleTwoDimensionalArray_Sort() {
	testCases := [][][]int{
		{{1, 1, 1}, {2, 3, 4}, {2, 6, 7}, {3, 4, 5}},
		{{1, 1, 1}},
		{{1, 1, 1}, {1, 1, 1}},
		{{1, 1, 1}, {2, 2, 3}, {2, 3, 3}, {1, 3, 4}},
	}

	for _, testCase := range testCases {
		tc := TwoDimensionalArray(testCase)
		sort.Sort(tc)
		fmt.Println(tc)
	}

	//output:
}
