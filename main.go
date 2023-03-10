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
