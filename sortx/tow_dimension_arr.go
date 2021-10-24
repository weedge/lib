package sortx

type TwoDimensionalArray [][]int

// sort by column asc
func (e TwoDimensionalArray) Less(i, j int) bool {
	for k := 0; k < len(e[i]); k++ {
		if e[i][k] < e[j][k] {
			return true
		} else if e[i][k] == e[j][k] {
			continue
		} else {
			return false
		}
	}
	return true
}

func (e TwoDimensionalArray) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
	return
}

func (e TwoDimensionalArray) Len() int {
	return len(e)
}
