package copy

func CopyTwoDimensionInt(src [][]int) [][]int {
	n := len(src)
	res := make([][]int, n)
	for i := 0; i < n; i++ {
		res[i] = make([]int, len(src[0]))
		copy(res[i], src[i])
	}

	return res
}

func CopyTwoDimensionInt64(src [][]int64) [][]int64 {
	n := len(src)
	res := make([][]int64, n)
	for i := 0; i < n; i++ {
		res[i] = make([]int64, len(src[0]))
		copy(res[i], src[i])
	}

	return res
}
