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
	watchChan := client.Watch(context.Background(), "name")
	for wresp := range watchChan {
		for _, evt := range wresp.Events {
			fmt.Printf("type:%s, key:%s, value: %s\n", evt.Type, evt.Kv.Key, evt.Kv.Value)
		}
	}

}
