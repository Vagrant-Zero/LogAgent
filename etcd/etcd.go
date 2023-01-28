package etcd

import (
	"context"
	"encoding/json"
	"github.com/coreos/etcd/clientv3"
	"github.com/sirupsen/logrus"
	"gocode/LogAgent_etcd/model"
	"gocode/LogAgent_etcd/tailfile"
	"time"
)

var (
	client *clientv3.Client
)

// NewEtcdClient 获取etcd客户端
func NewEtcdClient(address []string) (err error) {
	client, err = clientv3.New(clientv3.Config{
		Endpoints:   address,
		DialTimeout: time.Second * 10,
	})
	if err != nil {
		return err
	}
	logrus.Infof("init etcd client success!")
	return nil
}

// GetConfig 拉取日志收集项
func GetConfig(key string) (collectEntryList []model.CollectEntry) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	resp, err := client.Get(ctx, key)
	if err != nil {
		logrus.Errorf("get key = %s config from etcd failed, err = %v", key, err.Error())
		return nil
	}
	if len(resp.Kvs) == 0 {
		logrus.Warningf("length of key = %s config from etcd is 0", key)
		return nil
	}
	res := resp.Kvs[0]
	err = json.Unmarshal(res.Value, &collectEntryList)
	if err != nil {
		logrus.Errorf("unmarshal collectEntry failed, err = %v", err.Error())
		return nil
	}
	return collectEntryList
}

// WatchConfig 监控日志项：是否发生改变
func WatchConfig(key string) {
	// 循环监听配置是否发生改变
	for {
		watchOnce(key)
	}
}

func watchOnce(key string) {
	watchChan := client.Watch(context.Background(), key)
	var newConfig []model.CollectEntry
	for watchResp := range watchChan {
		logrus.Info("config has changed!")
		for _, evt := range watchResp.Events {
			err := json.Unmarshal(evt.Kv.Value, &newConfig)
			if err != nil {
				logrus.Errorf("new config unmarshal failed, err = %v", err.Error())
				continue
			}
			logrus.Infof("new config is {type:%s, key:%s, value: %s}", evt.Type, evt.Kv.Key, evt.Kv.Value)
			// 将新配置发往tailFile模块，更改配置
			tailfile.PutCollectEntry(newConfig)
		}
	}
}
