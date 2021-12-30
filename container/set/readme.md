#### 介绍

set 集合， 包括mapset ,  bitset(或者bitmap) ,  这里的hashset是以mapset来实现； 

#### 使用场景

其实主要都是在 **存在性** 问题 场景下的解决方案，尽量以低的空间利用率来解决问题，比如：

1. mapset 使用在某个元素是否已经在集合中了，map key是否存在；
2. bitset 使用在多个集合的diff, 并，与，异或等运算，计算汉明重量，以及偏移运算场景, 以及对bitset进行压缩([roaringbitmap](http://roaringbitmap.org/))；
2. Bitset 使用在类似0-1背包的问题，用于计算存放服用是否状态，进行状态转移，在高空间复杂度的情况下，优化内存消耗；
3. 由bitset衍生的bloom filter  过滤器，用于不存在的场景；

#### references

1. https://github.com/deckarep/golang-set
2. https://github.com/yourbasic/bit
2. https://github.com/bits-and-blooms/bitset
2. https://github.com/RoaringBitmap/roaring
2. [Consistently faster and smaller compressed bitmaps with Roaring.pdf](https://arxiv.org/pdf/1603.06549.pdf)
2. [An Experimental Study of Bitmap Compression vs. Inverted List Compression.pdf](https://w6113.github.io/files/papers/sidm338-wangA.pdf)
2. https://github.com/bits-and-blooms/bloom
2. [扶苏的bitset浅谈](https://www.cnblogs.com/yifusuyi/p/10072729.html)

