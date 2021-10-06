package strings

func ExampleString() {
	testBytes := []byte{'1', '2', '3'}
	s := String(testBytes)
	println(s, len(s))
	s1 := BytesToString(testBytes)
	println(s1, len(s1))
	// output:
	//
}
