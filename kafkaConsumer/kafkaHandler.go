package kafkaProducer

import (
	"db-consumer/common"
	"db-consumer/logger"
	"db-consumer/schema"
	"encoding/json"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type Producer struct {
	Pro  *kafka.Producer
	MapTopic *map[string]chan interface{}
}

func InitKafka() *Producer {
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": "localhost:9092"})
	if err != nil {
		panic(err)
	}

	// Delivery report handler for produced messages
	go func() {
		for e := range p.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					logger.LogError(common.SEND_MESSAGE_FAIL, fmt.Sprintf("%v \n%v", ev.TopicPartition, string(ev.Value)))
				} else {
					logger.LogInfo(common.SEND_MESSAGE, string(ev.Value))
				}
			}
		}
	}()
	mapTopic := make(map[string]chan interface{})
	return &Producer{
		Pro: p,
		MapTopic: &mapTopic,
	}
}

func (pro *Producer) ProcessSendMsg(schema *schema.Schema, size int)  {
	for name, _ := range schema.MapInfo {
		(*pro.MapTopic)[name] = make(chan interface{}, size)
		go func(topic string) {
			// Produce messages to topic (asynchronously)
			for {
				select {
				case msg := <- (*pro.MapTopic)[topic]:
					fmt.Println(msg)
					data, err := json.Marshal(msg)
					if err != nil {
						logger.LogError(fmt.Sprintf("%v", msg), err.Error())
					}
					pro.Pro.Produce(&kafka.Message{
						TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
						Value:          data,
					}, nil)
				}

			}
		}(name)
	}
}