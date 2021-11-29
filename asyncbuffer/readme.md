### intro
累积数据至buffer中，异步批量处理

#### 使用场景

1. Write-behind cache 模式，缓存批量入库，减少网络io, 以及入库磁盘io （补偿机制可以使用WAL(Write-Ahead Logging)的方式顺序写日志,todo）；对于写频率高的场景非常适合，比如投票，股票价格变动，以及课中直播互动场景等


#### 配置
```json
    {
      "async_buffer": {//用于异步批量处理的buffer配置列表
        "nmq_live_user": {//buffer名称
            "buffer_win_size": 1,//buffer窗口大小，等于这个值时触发批量操作
            "delay_do_ms": 10,//延时10ms做批量处理
            "chs": {//多个channel列表
                "nmq": {//channel名称
                    "ch_len": 0,//channel的长度
                    "sub_worker_num": 1//从channel中取数据的协程数目
                }
            }
        },
        "nmq_live_room": {
            "buffer_win_size": 1,
            "delay_do_ms": 10,
            "chs": {
                "nmq": {
                    "ch_len": 0,
                    "sub_worker_num": 1
                }
            }
        },
        "nmq_live_org": {
            "buffer_win_size": 1,
            "delay_do_ms": 10,
            "chs": {
                "nmq": {
                    "ch_len": 0,
                    "sub_worker_num": 1
                }
            }
        },
        "default": {
            "buffer_win_size": 1,
            "delay_do_ms": 10,
            "chs": {
                "default": {
                    "ch_len": 0,
                    "sub_worker_num": 1
                }
            }
        }
    }
  }

```

#### 流程

```
                                  buffer_win_size
                         ---ch1---  [======]  -sub batch worker-
IBuffer -(FormatInput)-> ---ch2---  [======]  -sub batch worker-  -(BatchDo)->(mq,cache,db)
                         ---ch3---  [======]  -sub batch worker-
                         ....                 .....
                     ---flushch---  [======]  -flush worker-

```

#### 使用
```go
type IBuffer interface {
	BatchDo([][]byte)//批量输出
	FormatInput() (error, []byte)//格式化输入的数据写入buffer
}
```
 1. 自定义根据buffer的实体数据结构, 实现IBuffer中的方法
 2. 通过 `SendOneCh(userBufferName, chName, data)` 方法将数据data写入对应的channel buffer中
 3. 通过`FlushAll();  FlushOne(bufferName string)` flush全部buffer,flush某个buffer(异步方式)

#### notice
 1. if batchDo panic, bufferData ingore

#### todo
- [ ] one IBuffer send to multi ch pub and sub batchDo 
- [ ] use multi buffer for multi sub worker to replace mutex lock



#### reference

1. [什么是 WAL](https://segmentfault.com/a/1190000022512468)
2. [Write-ahead_logging](https://en.wikipedia.org/wiki/Write-ahead_logging)
2. [ARIES:Algorithms_for_Recovery_and_Isolation_Exploiting_Semantics](https://en.wikipedia.org/wiki/Algorithms_for_Recovery_and_Isolation_Exploiting_Semantics)
2. [ARIES Overview, Types of Log Records, ARIES Helper Structures](https://www.youtube.com/watch?v=S9nctHdkggk)
2. [ARIES Database Recovery (CMU Databases Systems / Fall 2019)](https://www.youtube.com/watch?v=4VGkRXVM5fk)
2. [本地事务](http://icyfenix.cn/architect-perspective/general-architecture/transaction/local.html)

