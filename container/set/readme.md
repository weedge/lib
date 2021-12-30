#### 介绍

set 集合， 包括mapset ,  bitset(或者bitmap) ,  这里的hashset是以mapset来实现

#### 使用场景

其实主要都是在 **存在性** 问题 场景下的解决方案，尽量以低的空间利用率来解决问题，比如：

1. mapset 使用在某个元素是否已经在集合中了，map key是否存在
2. bitset 使用在多个集合的diff, 并，于，异或等运算，以及偏移运算场景, 以及对bitset进行压缩(roaringbitmap)，
3. 由bitset衍生的bloom filter  过滤器，用于不存在的场景

#### references

1. https://github.com/deckarep/golang-set
2. https://github.com/yourbasic/bit
2. https://github.com/bits-and-blooms/bitset
2. https://github.com/RoaringBitmap/roaring
2. https://github.com/bits-and-blooms/bloom

