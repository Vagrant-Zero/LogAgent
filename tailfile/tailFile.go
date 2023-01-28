package tailfile

import (
	"github.com/Shopify/sarama"
	"github.com/hpcloud/tail"
	"github.com/sirupsen/logrus"
	"gocode/LogAgent_etcd/kafka"
	"gocode/LogAgent_etcd/model"
	"strings"
	"time"
)

// 创建单个tailFile实例
func newTailFile(fileName string) (*tail.Tail, error) {
	config := tail.Config{
		ReOpen:    true,
		Follow:    true,
		Location:  &tail.SeekInfo{Offset: 0, Whence: 2},
		MustExist: false,
		Poll:      true,
	}
	// 打开文件
	tailFile, err := tail.TailFile(fileName, config)
	if err != nil {
		return nil, err
	}
	return tailFile, nil
}

// 读取日志，并将日志写入到msgChan中
func run(tailTask *model.TailTask) {
	var (
		msg      *tail.Line
		ok       bool
		kafkaMsg *sarama.ProducerMessage
	)
	for {
		select {
		case <-tailTask.Ctx.Done(): // 通知退出当前协程
			logrus.Infof("the goroutine named : %s has exist.", tailTask.TailFile.Filename)
			return
		case msg, ok = <-tailTask.TailFile.Lines:
			if !ok {
				logrus.Info("tail file close reopen , filename : ", tailTask.TailFile.Filename)
				time.Sleep(time.Second)
				continue
			}
			// 如果读到空行就跳过
			if len(strings.Trim(msg.Text, "\r")) == 0 {
				continue
			}
			// logrus.Info("读取到的msg为: ", msg.Text)
			// 将数据构造为kafka消息，并放入msgChan中
			kafkaMsg = &sarama.ProducerMessage{}
			kafkaMsg.Topic = tailTask.Topic
			kafkaMsg.Value = sarama.StringEncoder(msg.Text)
			// logrus.Infof("构造的kafka消息为: %#v", kafkaMsg)
			kafka.PutKafkaMsg(kafkaMsg)
		}
	}
}
