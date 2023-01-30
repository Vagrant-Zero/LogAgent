package main

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/sirupsen/logrus"
	"time"
)

func main() {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: time.Second * 10,
	})
	if err != nil {
		logrus.Errorf("failed to connect etcd, err = %v", err.Error())
		return
	}
	defer client.Close()
	fmt.Println(client)
	ip := "collectHostIP:"
	key := ip + "all_logs_config"
	value1 := "[{\"path\":\"D:\\\\Go_Learing\\\\src\\\\gocode\\\\LogAgent_etcd\\\\logFiles\\\\event_logs.log\",\"topic\":\"event_log\"}]\n"
	// value2 := "[{\"path\":\"D:\\\\Go_Learing\\\\src\\\\gocode\\\\LogAgent_etcd\\\\logFiles\\\\event_logs.log\",\"topic\":\"event_log\"},{\"path\":\"D:\\\\Go_Learing\\\\src\\\\gocode\\\\LogAgent_etcd\\\\logFiles\\\\mall_logs.log\",\"topic\":\"mall_log\"}]\n"
	// value3 := "[{\"path\":\"D:\\\\Go_Learing\\\\src\\\\gocode\\\\LogAgent_etcd\\\\logFiles\\\\event_logs.log\",\"topic\":\"event_log\"},{\"path\":\"D:\\\\Go_Learing\\\\src\\\\gocode\\\\LogAgent_etcd\\\\logFiles\\\\mall_logs.log\",\"topic\":\"mall_log\"},{\"path\":\"D:\\\\Go_Learing\\\\src\\\\gocode\\\\LogAgent_etcd\\\\logFiles\\\\shop_logs.log\",\"topic\":\"shop_log\"}]\n"
	// value4 := "[{\"path\":\"D:\\\\Go_Learing\\\\src\\\\gocode\\\\LogAgent_etcd\\\\logFiles\\\\event_logs.log\",\"topic\":\"event_log\"},{\"path\":\"D:\\\\Go_Learing\\\\src\\\\gocode\\\\LogAgent_etcd\\\\logFiles\\\\mall_logs.log\",\"topic\":\"mall_log\"},{\"path\":\"D:\\\\Go_Learing\\\\src\\\\gocode\\\\LogAgent_etcd\\\\logFiles\\\\shop_logs.log\",\"topic\":\"shop_log\"},{\"path\":\"D:\\\\Go_Learing\\\\src\\\\gocode\\\\LogAgent_etcd\\\\logFiles\\\\account_logs.log\",\"topic\":\"account_log\"}]\n"
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	pair, err := client.Put(ctx, key, value1)
	cancel()
	if err != nil {
		logrus.Errorf("failed to put kv pair etcd, err = %v", err.Error())
	}
	logrus.Infof("pair = %v", pair)
	ctx1, cancel1 := context.WithTimeout(context.Background(), time.Second)
	getResponse, err := client.Get(ctx1, key)
	if err != nil {
		logrus.Errorf("get kv pair failed from etcd, err = %v", err.Error())
	}
	cancel1()
	for _, kv := range getResponse.Kvs {
		logrus.Infof("key = %v, value = %v", string(kv.Key), string(kv.Value))
	}
}
