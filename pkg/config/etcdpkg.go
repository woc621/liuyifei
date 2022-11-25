package config

import (
	"context"
	"fmt"
	"liuyifei/pkg/util"
	"time"

	etcdClient "go.etcd.io/etcd/clientv3"
)

func init() {
	//cfgfile = ""
}

const (
	dialtimeoutunits       = time.Second
	DEFAULT_PLATFORM_VALUE = "system"
	INI_CFG_PLATFORM       = "platform"
	INI_CFG_PLATFORM_key   = "platform"
)

type LogConfig struct {
	PlatForm string
	LogPath  map[string]string
}

func GetClientEtcd(cfg CommonConfig) (etcdcli *etcdClient.Client, err error) {
	etcdcfg := etcdClient.Config{
		Endpoints:   cfg.Endpoints,
		DialTimeout: time.Duration(cfg.DialTimeout) * time.Second,
	}
	etcdcli, err = etcdClient.New(etcdcfg)
	return
}

func (cfg CommonConfig) GetLogCfg(etcdCli *etcdClient.Client) (logcfg LogConfig, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	prefix := fmt.Sprintf("/%s", cfg.Platform)
	resp, err := etcdCli.Get(ctx, prefix, etcdClient.WithPrefix())
	defer cancel()
	if err != nil {
		fmt.Println("get etcd err:", err)
		return
	}
	logcfg.LogPath =make(map[string]string)
	logcfg.PlatForm = cfg.Platform
	for _, kv := range resp.Kvs {
		//判断获取的日志路径是否是目录，如果是目录就跳过这次循环不加载。
		if util.IsDir(string(kv.Value)) {
			fmt.Printf("%s 是一个目录,请重新配置\n", string(kv.Value))
			continue
		}
		key := fmt.Sprintf("%s", kv.Key)
		value := fmt.Sprintf("%s", kv.Value)
		logcfg.LogPath[key] = value
	}
	return
}
