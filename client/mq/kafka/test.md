#### 安装测试:

开发环境： mac

安装：brew install kafka

开箱即用，启动：

```shell
# 启动zookeeper，单个服务，
/usr/local/bin/zookeeper-server-start /usr/local/etc/kafka/zookeeper.properties
# 启动kafka broker 0
/usr/local/bin/kafka-server-start /usr/local/etc/kafka/server.properties
# 启动kafka broker 1 修改broker.id=1 listeners=PLAINTEXT://:9093 log.dirs=/usr/local/var/lib/kafka-1-logs
/usr/local/bin/kafka-server-start /usr/local/etc/kafka/server-1.properties
# 启动kafka broker 2 修改broker.id=2 listeners=PLAINTEXT://:9094 log.dirs=/usr/local/var/lib/kafka-2-logs
/usr/local/bin/kafka-server-start /usr/local/etc/kafka/server-2.properties
# 创建topic, topic 分片数，副本数：为1只有一个主副本，没有备份
/usr/local/bin/kafka-topics  --create --zookeeper localhost:2181 --replication-factor 3 --partitions 2 --topic sarama
# 对sarama topic生产数据
/usr/local/bin/kafka-console-producer --broker-list localhost:9092 --topic sarama
# 从开始offset 消费 sarama topic
/usr/local/bin/kafka-console-consumer --bootstrap-server localhost:9092 --topic sarama --from-beginning
```

#####  /usr/local/etc/kafka/server.properties 配置:

```shell
#查看kafka.server.KafkaConfig获取更多细节和默认值

############################# 服务器基础配置 Server Basics #############################

#代理的id。对于每个代理，这必须设置为唯一的整数。
broker.id = 0

############################# 套接字服务器设置  Socket Server Settings 网络模型设置#############################

#socket服务器监听的地址。它将得到返回的值
# java.net.InetAddress.getCanonicalHostName()如果没有配置。
#格式:
# listener = listener_name://host_name:port
#例子:
# listener = PLAINTEXT://your.host.name:9092
listeners=PLAINTEXT://:9092

#代理将向生产者和消费者发布的主机名和端口。
#如果配置了监听器，则使用监听器的值。否则，它将使用该值
#从java.net.InetAddress.getCanonicalHostName()返回。
#advertised.listeners=PLAINTEXT://your.host.name:9092

#将监听器名称映射到安全协议，默认情况下它们是相同的。有关更多细节，请参阅配置文档
#listener.security.protocol.map=PLAINTEXT:PLAINTEXT,SSL:SSL,SASL_PLAINTEXT:SASL_PLAINTEXT,SASL_SSL:SASL_SSL

#服务器从网络接收请求并向网络发送响应的线程数
num.network.threads = 3

#服务器用于处理请求的线程数，可能包括磁盘I/O
num.io.threads = 8

#socket服务器使用的发送缓冲区 SO_SNDBUF
socket.send.buffer.bytes = 102400

#socket服务器使用的接收缓冲区 SO_RCVBUF
socket.receive.buffer.bytes = 102400

#socket服务器将接受的最大请求大小(针对OOM的保护)
socket.request.max.bytes = 104857600


############################# 日志基础配置 Log Basics ##############################

#以逗号分隔的目录列表，用于存储日志文件
log.dirs=/usr/local/var/lib/kafka-logs

# 每个主题的默认日志分区数。更多分区允许更大分区
# parallelism for consumption，但这也会导致更多的文件在broker上。
num.partitions = 3

#每个数据目录用于启动时日志恢复和关闭时刷新的线程数。
#对于数据目录位于RAID的安装，建议增加该值。
num.recovery.threads.per.data.dir = 1

############################# 内部主题设置   #############################
#组元数据内部主题“__consumer_offset”和“__transaction_state”的复制因子
#对于开发测试以外的任何情况，建议使用大于1的值，以确保可用性，例如3。
offsets.topic.replication.factor = 1
transaction.state.log.replication.factor = 1
transaction.state.log.min.isr = 1

############################# 日志刷新策略  Log Flush Policy #############################

#消息立即写入文件系统，但默认情况下，我们只使用fsync()进行同步
#操作系统缓存延迟。以下配置控制将数据刷新到磁盘。
#这里有一些重要的权衡:
# 1。持久性:如果不使用复制，未刷新的数据可能会丢失。
# 2。延迟:当确实发生刷新时，非常大的刷新间隔可能会导致延迟尖峰，因为有大量数据要刷新。
# 3。吞吐量:刷新通常是开销最大的操作，较小的刷新间隔可能导致过多的搜索。
#下面的设置允许将刷新策略配置为在一段时间或之后刷新数据
#每个N条消息(或两者)。这可以全局完成，并在每个主题的基础上重写。

#在强制将数据刷新到磁盘之前接受的消息数
# log.flush.interval.messages = 10000

#在强制刷新之前，消息可以在日志中保存的最大时间
# log.flush.interval.ms = 1000

############################# 日志保留策略  Log Retention Policy #############################

#以下配置控制日志段的处理。政策可以
#被设置为在一段时间后，或在给定的大小积累后删除段。
#当满足这些条件时，段将被删除。删除总是发生
#从日志的末尾。

#日志文件的最小年龄可以被删除
log.retention.hours = 168

#基于大小的日志保留策略。除非有剩余的片段，否则将从日志中删除
# segment drop below log.retention.bytes。功能独立于日志。保留。小时。
# log.retention.bytes = 1073741824

#日志段文件的最大大小。当达到这个大小时，将创建一个新的日志段。
log.segment.bytes = 1073741824

#检查日志段的时间间隔，查看是否可以根据该时间间隔删除日志段
#保留政策
log.retention.check.interval.ms = 300000

############################# zookeeper  #############################

# Zookeeper连接字符串(详见Zookeeper文档)。
# 这是一个逗号分隔的主机:端口对，每个对应一个zk
#服务器。如。“127.0.0.1:3000 127.0.0.1:3001 127.0.0.1:3002”。
#你也可以附加一个可选的chroot字符串到url来指定
#所有kafka znodes的根目录。
zookeeper.connect = localhost: 2181

#连接zookeeper超时，单位为ms
zookeeper.connection.timeout.ms = 18000


############################# 组织协调器设置  Group Coordinator Settings  #############################

#下面的配置指定了GroupCoordinator延迟初始消费者重新平衡的时间，以毫秒为单位。
#当新成员加入组时，再平衡将被进一步延迟group.initial.rebalance.delay.ms值，最大延迟为max.poll.interval.ms。
#默认值是3秒。
#我们在这里将其重写为0，因为这有利于开发和测试的开箱即用体验。
#然而，在生产环境中，默认值3秒更合适，因为这将有助于避免在应用程序启动期间不必要的、可能昂贵的重新平衡。
group.initial.rebalance.delay.ms = 0

```

