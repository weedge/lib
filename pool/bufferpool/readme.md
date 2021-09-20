##### 介绍

通过sync.Pool池化复用buffer, 使用场景：大量临时对象append操作，编码解码操作等，通过复用减少gc

##### 对比

```shell
go test -bench=. -run=none  -benchmem=1
goos: darwin
goarch: amd64
pkg: github.com/weedge/lib/pool/bufferpool
Benchmark_TestStringAppend-8              	  328002	      3629 ns/op	   28672 B/op	       4 allocs/op
Benchmark_TestBufferPool-8                	 7447075	       157 ns/op	       0 B/op	       0 allocs/op
Benchmark_TestBuffer-8                    	  300680	      3643 ns/op	   25216 B/op	       3 allocs/op
BenchmarkJI_TestBufferPool_Parallel-8     	32020629	        34.6 ns/op	       0 B/op	       0 allocs/op
BenchmarkJI_TestBuffer_Parallel-8         	  315584	      3517 ns/op	   25216 B/op	       3 allocs/op
BenchmarkJI_TestStringAppend_Parallel-8   	  271797	      4231 ns/op	   28672 B/op	       4 allocs/op
```

总结： 使用池化buffer, 减少了对象分配次数，减少gc