### intro
累积数据至buffer中，异步批量处理

#### 使用场景

1. Write-behind cache 模式，缓存批量入库，减少网络io, 以及入库磁盘io （补偿机制可以使用WAL(Write-Ahead Logging)的方式顺序写日志）；对于写频率高的场景非常适合，比如投票，股票价格变动，以及课中直播互动场景等

   `在 write-behind 缓存中，数据的读取和更新通过缓存进行，与 write-through 缓存不同，更新的数据并不会立即传到数据库。相反，在缓存中一旦进行更新操作，缓存就会跟踪脏记录列表，并定期将当前的脏记录集刷新到数据库中。作为额外的性能改善，缓存会合并这些脏记录。合并意味着如果相同的记录被更新，或者在缓冲区内被多次标记为脏数据，则只保证最后一次更新。对于那些值更新非常频繁，例如金融市场中的股票价格等场景，这种方式能够很大程度上改善性能。如果股票价格每秒钟变化 100 次，则意味着在 30 秒内会发生 30 x 100 次更新。合并将其减少至只有一次。`

   `Write-behind 缓存并不能放之四海而皆准。write-behind 的本质注定用户看到的变化，即使被提交也不会立即反映到数据库中。这种时间延迟被称为“缓存写延迟（cache write latency）”或“数据库腐败（database staleness）”；发生在数据库变更与更新数据（或者使得数据无效）以反映其变更的缓存之间的延迟则被称为“缓存读延迟（cache read latency）”或“缓存腐败（cache staleness）”。如果系统的每部分在访问数据时都通过缓存（例如，通过公共接口），那么，由于缓存总是保持最新的正确记录，采用 write-behind 技术就是值得的。可以预见，采用 write-behind 的系统，作出变更的唯一路径就只能是缓存。`

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
2. [极端事务处理模式：Write-behind 缓存](https://www.infoq.cn/article/write-behind-caching)

