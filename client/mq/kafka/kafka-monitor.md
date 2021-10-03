监控环境： mac

```shell
# install prometheus
brew install prometheus

# install grafana
brew install grafana

# install kafka_exporter
wget https://github.com/danielqsj/kafka_exporter/releases/download/v1.3.1/kafka_exporter-1.3.1.darwin-amd64.tar.gz -P /usr/local/Cellar/kafka_exporter && cd /usr/local/Cellar/kafka_exporter
tar -xzvf kafka_exporter-1.3.1.darwin-amd64.tar.gz 


# start kafka_exporter job   http://localhost:9308
./kafka_exporter --kafka.server=localhost:9092 --kafka.server=localhost:9093 --kafka.server=localhost:9094

# 修改 prometheus.yml，加上 kafka_exporter 的 job。默认端口是 9308。
cat /usr/local/etc/prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: "prometheus"
    static_configs:
    - targets: ["localhost:9090"]
  - job_name: "kafka_exporter"
    static_configs:
    - targets: ["localhost:9308"]
    
# start prometheus http://localhost:9090
brew services start prometheus

# start grafana http://localhost:3000
brew services start grafana


# import Grafana Dashboard ID: 7589, name: Kafka Exporter Overview.

# config panel add metrics
# view Grafana Dashboard monitor broker, topic, partition offset/replicas/isr, consumergroup offset/lag ...etc
```
##### 监控指标

| Metrics                                          | Description                               | 维度指标实例(from Prometheus)                                |
| ------------------------------------------------ | ----------------------------------------- | ------------------------------------------------------------ |
| kafka_brokers                                    | kafka 集群的 broker 数量                  | kafka_brokers{**instance**="localhost:9308", **job**="kafka_exporter"} |
| kafka_topic_partitions                           | kafka topic 的分区数                      | kafka_topic_partitions{**instance**="localhost:9308", **job**="kafka_exporter", **topic**="sarama"} |
| kafka_topic_partition_current_offset             | kafka topic 分区当前的 offset             | kafka_topic_partition_current_offset{**instance**="localhost:9308", **job**="kafka_exporter", **partition**="0", **topic**="sarama"} |
| kafka_topic_partition_oldest_offset              | kafka topic 分区最旧的 offset             | kafka_topic_partition_oldest_offset{**instance**="localhost:9308", **job**="kafka_exporter", **partition**="0", **topic**="sarama"} |
| kafka_topic_partition_in_sync_replica            | 处于同步过程中的 topic/partition 数       | kafka_topic_partition_in_sync_replica{**instance**="localhost:9308", **job**="kafka_exporter", **partition**="0", **topic**="sarama"} |
| kafka_topic_partition_leader                     | topic/partition leader 的 broker id       | kafka_topic_partition_leader{**instance**="localhost:9308", **job**="kafka_exporter", **partition**="0", **topic**="sarama"} |
| kafka_topic_partition_leader_is_preferred        | topic/partition 是否使用 preferred broker | kafka_topic_partition_leader_is_preferred{**instance**="localhost:9308", **job**="kafka_exporter", **partition**="0", **topic**="sarama"} |
| kafka_topic_partition_replicas                   | topic/partition 的副本数                  | kafka_topic_partition_replicas{**instance**="localhost:9308", **job**="kafka_exporter", **partition**="0", **topic**="sarama"} |
| kafka_topic_partition_under_replicated_partition | partition 是否处于 replicated             | kafka_topic_partition_under_replicated_partition{**instance**="localhost:9308", **job**="kafka_exporter", **partition**="0", **topic**="sarama"} |
| kafka_consumergroup_current_offset               | kakfa topic 消费者组的 offset             | kafka_consumergroup_current_offset{**consumergroup**="consumer.group.test", **instance**="localhost:9308", **job**="kafka_exporter", **partition**="0", **topic**="sarama"} |
| kafka_consumergroup_current_offset_sum           | kakfa topic 消费者组的 offset 总数        | kafka_consumergroup_current_offset_sum{**consumergroup**="consumer.group.test", **instance**="localhost:9308", **job**="kafka_exporter", **topic**="sarama"} |
| kafka_consumergroup_lag_sum                      | kakfa-lag 消费延迟 总数                   | kafka_consumergroup_lag_sum{**consumergroup**="consumer.group.test", **instance**="localhost:9308", **job**="kafka_exporter", **topic**="sarama"} |
| kafka_consumergroup_lag                          | kakfa-lag 消费延迟                        | kafka_consumergroup_lag{**consumergroup**="consumer.group.test", **instance**="localhost:9308", **job**="kafka_exporter", **partition**="0", **topic**="sarama"} |
| kafka_consumergroup_members                      | kakfa topic 消费者组成员                  | kafka_consumergroup_members{**consumergroup**="consumer.group.test", **instance**="localhost:9308", **job**="kafka_exporter"} |



##### reference

1. [Prometheus+Grafana+kafka_exporter搭建监控系统监控kafka](https://github.com/Lancger/opslinux/blob/master/kafka/Prometheus%2BGrafana%2Bkafka_exporter%E6%90%AD%E5%BB%BA%E7%9B%91%E6%8E%A7%E7%B3%BB%E7%BB%9F%E7%9B%91%E6%8E%A7kafka.md)