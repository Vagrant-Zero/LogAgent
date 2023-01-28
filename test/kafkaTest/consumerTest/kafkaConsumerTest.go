package main

import (
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/sirupsen/logrus"
	"sync"
)

func main() {
	consumer, err := sarama.NewConsumer([]string{"127.0.0.1:9092"}, nil)
	if err != nil {
		logrus.Errorf("failed to connect kafka consumer, err = %v", err.Error())
		return
	}
	defer consumer.Close()
	// 获取topic下的所有分区列表
	partitionList, err := consumer.Partitions("account_log")
	if err != nil {
		logrus.Errorf("failed to get kafka partitions, err = %v", err.Error())
		return
	}
	var wg sync.WaitGroup
	// 遍历所有分区
	for partition := range partitionList {
		// 对每个分区创建一个消费者
		partitionConsumer, err := consumer.ConsumePartition("account_log", int32(partition), sarama.OffsetNewest)
		if err != nil {
			logrus.Errorf("failed to create kafka partition consumer, err = %v", err.Error())
			return
		}
		defer partitionConsumer.AsyncClose()
		// 异步从每个分区消费数据
		wg.Add(1)
		go func(pc sarama.PartitionConsumer) {
			for msg := range pc.Messages() {
				fmt.Printf("Partition: %d, Offset: %d, key: %s, value: %s\n",
					msg.Partition, msg.Offset, msg.Key, msg.Value)
			}
		}(partitionConsumer)
	}
	wg.Wait() // 等待其他的goroutine任务完毕
}
