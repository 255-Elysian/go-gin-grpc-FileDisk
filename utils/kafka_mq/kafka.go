package kafka_mq

import (
	"github.com/segmentio/kafka-go"
	"grpc-todolist-disk/conf"
)

// NewKafkaConsumer 创建 Kafka 消费者
func NewKafkaConsumer() *kafka.Reader {
	// 创建 Kafka reader
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   conf.Conf.Kafka.Broker,  // Kafka 集群地址
		Topic:     conf.Conf.Kafka.Topic,   // 要订阅的主题
		GroupID:   conf.Conf.Kafka.GroupId, // 消费组 ID（用于负载均衡和 Offset 管理）
		Partition: 0,                       // 可省略，设置 GroupID 时 kafka-go 会自动处理分区
		MinBytes:  10e3,                    // 10KB 最小获取消息体积
		MaxBytes:  10e6,                    // 10MB 最大获取消息体积
	})
	return r
}

// NewKafkaProducer 创建一个 Kafka 生产者
func NewKafkaProducer() *kafka.Writer {
	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers: conf.Conf.Kafka.Broker,
		Topic:   conf.Conf.Kafka.Topic,
	})
	return w
}
