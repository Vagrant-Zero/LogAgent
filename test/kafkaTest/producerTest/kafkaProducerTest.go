package main

import (
	"fmt"
	"github.com/Shopify/sarama"
	"log"
)

func main() {
	// 生产者配置
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll          //Ack
	config.Producer.Partitioner = sarama.NewRandomPartitioner //Partition
	config.Producer.Return.Successes = true                   // 发送消息成功的确认
	// 连接kafka
	client, err := sarama.NewSyncProducer([]string{"127.0.0.1:9092"}, config)
	if err != nil {
		log.Fatalf("连接到kafka生产者出错，err = %V", err.Error())
	}
	defer client.Close()
	// 发送消息
	msg := &sarama.ProducerMessage{}
	msg.Topic = "web_logs"
	msg.Value = sarama.StringEncoder("this is a test message!")
	// 发送消息
	pid, offst, err := client.SendMessage(msg)
	if err != nil {
		log.Fatalf("发送消息失败，err = %v", err.Error())
		return
	}
	fmt.Printf("pid: %v, offset: %v\n", pid, offst)
}
