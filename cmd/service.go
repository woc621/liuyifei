package cmd

import (
	"liuyifei/pkg/config"

	etcdClient "go.etcd.io/etcd/clientv3"
)
type ServiceConfig struct {
	Cfg     config.CommonConfig
	EtcdCli *etcdClient.Client
	CtxMap  map[string]config.CtxS
	KafkaProducer *config.KafkaProducer
	LogCfg  config.LogCfg
	config.SystemInfo
}
