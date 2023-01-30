# LogAgent
该项目是用于记录日志文件，并发往kafka进行消费，特点是可以动态的增加/删除监听的日志文件目录。该项目参考七米老师的实现，非常感谢七米老师的讲解！
## 1. etcd配置

etcd配置的需要收集日志的格式

```json
[
    {
    "path":"xxxx\logFiles\event_logs.log",
    "topic":"event_log"
    },
    {
    "path":"xxxx\logFiles\mall_logs.log",
    "topic":"mall_log"
    }
]
```

## 2. 收集多个日志文件

针对每个路径，配置一个`tailTask`，创建收集任务，并启动协程去执行

```go
package model

import (
	"context"
	"github.com/hpcloud/tail"
)

// TailTask tail任务对象
type TailTask struct {
	TailFile *tail.Tail
	Topic    string
	Ctx      context.Context
	Cancel   context.CancelFunc
}

func NewTailTask(tailFile *tail.Tail, topic string) *TailTask {
	ctx, cancel := context.WithCancel(context.Background())
	return &TailTask{
		TailFile: tailFile,
		Topic:    topic,
		Ctx:      ctx,
		Cancel:   cancel,
	}
}
```



## 3. 管理日志收集

+ 管理日志收集，即：新增或者删除target目录的日志时，需要新增或者删除读取日志，并发往kafka
+ 使用协程以和`context`进行处理，初始化任务的时候就开启后台协程去监听有无新配置的到来，如果有新配置，则通过`TailTask.Cancel`通知协程退出

```go
// TailTaskManager

// 查找老配置中存在，但是新配置中不存在的配置，并在老配置map中删除
func deleteExistTailTask(newConfigMap map[string]struct{}) {
	for key, _ := range tailTaskMap {
		if _, exist := newConfigMap[key]; !exist {
			// 老配置中不存在，删除对应tailTask
			logrus.Infof("the goroutine named : %s is ready to stop.", key)
			tailTask := tailTaskMap[key]
			delete(tailTaskMap, key)
			// 并且停掉对应的goroutine
			tailTask.Cancel() //结束tail.run()函数
		}
	}
}
```

```go
// TailFile

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
```

## 4. 根据IP获取配置

```go
package common

import (
	"github.com/sirupsen/logrus"
	"net"
	"strings"
)

const CanNotGetIp = "0x3f3f"

// GetOutBoundIP 获取本机IP地址
func GetOutBoundIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		defer conn.Close()
		return CanNotGetIp, err
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	ip := strings.Split(localAddr.IP.String(), ":")[0]
	logrus.Infof("get localhost IP success! ip = %s", ip)
	return ip, nil
}
```

## 5. 全局逻辑梳理

`main`函数中给出了全部流程，其中后台协程的启动在tailTask的初始化中，后台接受新配置

```go
package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"gocode/LogAgent_etcd/common"
	"gocode/LogAgent_etcd/etcd"
	"gocode/LogAgent_etcd/kafka"
	"gocode/LogAgent_etcd/tailfile"
	"gocode/LogAgent_etcd/utils"
	"sync"
)

func main() {
	// 0. 获取本机IP，为后续从etcd拉取配置做准备
	ip, err := common.GetOutBoundIP()
	if err != nil {
		logrus.Errorf("get local IP failed, err = %v", err.Error())
		return
	}

	// 1. 读取配文件
	cfg, err, done := utils.ReadConfig()
	if done {
		logrus.Error("failed to load config file, err = ", err.Error())
		return
	}

	// 2. 初始化kafka
	err = kafka.NewKafkaClient([]string{cfg.KafkaConfig.Address}, cfg.ChanSize)
	if err != nil {
		logrus.Error("kafka: failed to init kafka client, err = ", err.Error())
	}

	// 3. 初始化etcd连接
	err = etcd.NewEtcdClient([]string{cfg.EtcdConfig.Address})
	if err != nil {
		logrus.Errorf("init etcd failed, err = %v", err.Error())
		return
	}

	// 4. 从etcd中拉取配置
	collectKey := fmt.Sprintf(cfg.EtcdConfig.CollectKey, ip)
	// collectEntryList := etcd.GetConfig(cfg.EtcdConfig.CollectKey)
	collectEntryList := etcd.GetConfig(collectKey)

	// 4. 根据etcd中拉取的配置，初始化tail,读取日志文件
	err = tailfile.InitTailTaskList(collectEntryList)
	if err != nil {
		logrus.Errorf("failed to load log files，err = %v", err.Error())
	}

	// 启动协程，监控etcd中配置项是否改变
	// go etcd.WatchConfig(cfg.EtcdConfig.CollectKey)
	go etcd.WatchConfig(collectKey)

	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}
```
