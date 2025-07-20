package mq

import (
	"github.com/segmentio/kafka-go"
	"grpc-todolist-disk/utils/kafka_mq"
)

var KfWriter *kafka.Writer

func Init() {
	KfWriter = kafka_mq.NewFileKafkaProducer()
}
