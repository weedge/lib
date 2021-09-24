#### 介绍

golang的map结构遍历(range map)的时候是无序的，因为底层的map存放数据是在hash表中, 可以通过汇编代码查看到调用的地方:

- runtime.mapiterinit
- runtime.mapiternext

```assembly
	0x009b 00155 (main.go:11)	LEAQ	type.map[int32]string(SB), AX
	0x00a2 00162 (main.go:11)	PCDATA	$2, $0
	0x00a2 00162 (main.go:11)	MOVQ	AX, (SP)
	0x00a6 00166 (main.go:11)	PCDATA	$2, $2
	0x00a6 00166 (main.go:11)	LEAQ	""..autotmp_3+24(SP), AX
	0x00ab 00171 (main.go:11)	PCDATA	$2, $0
	0x00ab 00171 (main.go:11)	MOVQ	AX, 8(SP)
	0x00b0 00176 (main.go:11)	PCDATA	$2, $2
	0x00b0 00176 (main.go:11)	LEAQ	""..autotmp_2+72(SP), AX
	0x00b5 00181 (main.go:11)	PCDATA	$2, $0
	0x00b5 00181 (main.go:11)	MOVQ	AX, 16(SP)
	0x00ba 00186 (main.go:11)	CALL	runtime.mapiterinit(SB)
	0x00bf 00191 (main.go:11)	JMP	207
	0x00c1 00193 (main.go:11)	PCDATA	$2, $2
	0x00c1 00193 (main.go:11)	LEAQ	""..autotmp_2+72(SP), AX
	0x00c6 00198 (main.go:11)	PCDATA	$2, $0
	0x00c6 00198 (main.go:11)	MOVQ	AX, (SP)
	0x00ca 00202 (main.go:11)	CALL	runtime.mapiternext(SB)
	0x00cf 00207 (main.go:11)	CMPQ	""..autotmp_2+72(SP), $0
	0x00d5 00213 (main.go:11)	JNE	193
```

**runtime.mapiterinit**:

```go
...
// decide where to start
r := uintptr(fastrand())
if h.B > 31-bucketCntBits {
	r += uintptr(fastrand()) << 31
}
it.startBucket = r & bucketMask(h.B)
it.offset = uint8(r >> h.B & (bucketCnt - 1))

// iterator state
it.bucket = it.startBucket
....
```

无序的原因： <u>在开始处理循环逻辑的时候，就做了随机播种</u>

为了解决这个无序的问题，将map结构转换成slice结构，通过sort.Slice()来排序；

#### 功能

- [ ] 支持map[int64]int64 key/value的升/降排序: SortIntIntMapByValue, SortIntIntMapByValueDesc, SortIntIntMapByKey, SortIntIntMapByKeyDesc
- [ ] 支持map[int64]string key/value的升/降排序: SortIntStringMapByValue, SortIntStringMapByValueDesc, SortIntStringMapByKey, SortIntStringMapByKeyDesc
- [ ] 支持map[string]string key/value的升/降排序: SortStringStringMapByValue, SortStringStringMapByValueDesc, SortStringStringMapByKey, SortStringStringMapByKeyDesc
- [ ] 支持map[string]int64 key/value的升/降排序: SortStringIntMapByValue, SortStringIntMapByValueDesc, SortStringIntMapByKey, SortStringIntMapByKeyDesc

#### 总结

​	这个实现有点体力活，思路都是一样的，只是不同的map k/v类型 定义的 Len,Swap,Less 不同，后续go(1.17已经支持)支持模版了，用模版实现下

#### reference

1. [Go maps in action](https://go.dev/blog/maps)
2. [map源码解析](https://github.com/cch123/golang-notes/blob/master/map.md)

