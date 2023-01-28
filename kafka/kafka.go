package kafka

import (
	"github.com/Shopify/sarama"
	"github.com/sirupsen/logrus"
)

var (
	kafkaClient sarama.SyncProducer
	msgChan     chan *sarama.ProducerMessage
	err         error
)

// NewKafkaClient 单例模式获取kafkaClient
func NewKafkaClient(address []string, chanSize int) error {
	if kafkaClient != nil {
		return nil
	}
	msgChan = make(chan *sarama.ProducerMessage, chanSize)
	// 1. 生产者配置
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll          //Ack
	config.Producer.Partitioner = sarama.NewRandomPartitioner //Partition
	config.Producer.Return.Successes = true                   // 发送消息成功的确认
	// 2. 连接kafka，获取kafka生产者客户端
	kafkaClient, err = sarama.NewSyncProducer(address, config)
	if err != nil {
		return err
	}
	logrus.Info("init kafkaClient success!")
	go sendMsg()
	return nil
}

func sendMsg() {
	defer kafkaClient.Close()
	for {
		select {
		case msg := <-msgChan:
			pid, offset, err := kafkaClient.SendMessage(msg)
			if err != nil {
				logrus.Warn("send message failed，err = ", err.Error())
			} else {
				logrus.Infof("send message success! pid = %v, offset = %v", pid, offset)
			}
		}
	}
}

// PutKafkaMsg 将消息放入msgChan中，供后续发送消息到kafka
func PutKafkaMsg(kafkaMsg *sarama.ProducerMessage) {
	msgChan <- kafkaMsg
}
