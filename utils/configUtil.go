package utils

import (
	"gocode/LogAgent_etcd/model"
	"gopkg.in/ini.v1"
)

// ReadConfig 读取配置文件
func ReadConfig() (*model.Config, error, bool) {
	// 1. 读取配置文件
	cfg := new(model.Config)
	err := ini.MapTo(cfg, "./config/config.ini")
	if err != nil {
		return nil, nil, true
	}
	return cfg, err, false
}
