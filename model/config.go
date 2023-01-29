package model

// Config 全局配置文件
type Config struct {
	KafkaConfig `ini:"kafka"`
	// CollectConfig `ini:"collect"`
	EtcdConfig `ini:"etcd"`
}
