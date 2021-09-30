package strings

import (
	"strconv"
	"unsafe"
)

func BytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func isIntNum(str string) bool {
	_, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return false
	}

	return true
}

func isFloatNum(str string) bool {
	_, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return false
	}

	return true
}
