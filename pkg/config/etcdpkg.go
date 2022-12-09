package config

import (
	"context"
	"fmt"
	"strings"
	"time"

	etcdClient "go.etcd.io/etcd/clientv3"
)

func init() {
	//cfgfile = ""
}

const (
	dialtimeoutunits        = time.Second
	DEFAULT_MAC_VALUE       = "default"
	WINDOWS_DEFAULT_LOGNAME = "system"
	WINDOWS_DEFAULT_LOGFILE = "c:/log.log"
	LINUX_DEFAULT_LOGNAME   = "system"
	LINUX_DEFAULT_LOGFILE   = "/var/log/messages"
)

var (
	EtcdCli *etcdClient.Client
	Logcfg  LogCfg
)

type LogCfg struct {
	ClientName string
	PlatForm   string
	LogPath    map[string]string
}

func GetClientEtcd(cfg CommonConfig) (etcdcli *etcdClient.Client, err error) {
	etcdcfg := etcdClient.Config{
		Endpoints:   cfg.Endpoints,
		DialTimeout: time.Duration(cfg.DialTimeout) * time.Second,
	}
	etcdcli, err = etcdClient.New(etcdcfg)
	EtcdCli = etcdcli
	return
}
func GetLogCfgByPrefix(prefix string) (logcfg LogCfg, err error) {
	logcfg.ClientName = CommonCfg.Mac
	logcfg.PlatForm = CommonCfg.Platform
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	//etcd默认配置key:prefix
	// defaultprefix := fmt.Sprintf("/%s/%s/default", CommonCfg.Os,CommonCfg.Platform)
	//获取默认配置value
	resp, err := EtcdCli.Get(ctx, prefix, etcdClient.WithPrefix())
	defer cancel()
	if err != nil {
		fmt.Println("get etcd err:", err)
		return logcfg, err
	}
	logcfg.LogPath = make(map[string]string)
	for _, kv := range resp.Kvs {
		logcfg.LogPath[string(kv.Key)] = string(kv.Value)
	}

	return logcfg, nil
}
func SetDefaultLogCfg() {

	defaultprefix := fmt.Sprintf("/%s/%s/default", CommonCfg.Os, CommonCfg.Platform)
	logcfg, _ := GetLogCfgByPrefix(defaultprefix)

	if len(logcfg.LogPath) == 0 {
		fmt.Println("获取默认日志配置")
		if CommonCfg.Os == "windows" {
			key := fmt.Sprintf("/%s/%s/%s/%s", CommonCfg.Os, CommonCfg.Platform, CommonCfg.Mac, WINDOWS_DEFAULT_LOGNAME)
			logcfg.LogPath[key] = WINDOWS_DEFAULT_LOGFILE
		} else {
			key := fmt.Sprintf("/%s/%s/%s/%s", CommonCfg.Os, CommonCfg.Platform, CommonCfg.Mac, LINUX_DEFAULT_LOGNAME)
			logcfg.LogPath[key] = LINUX_DEFAULT_LOGFILE
		}
	}
	for key, value := range logcfg.LogPath {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		old := fmt.Sprintf("/%s/%s/%s", CommonCfg.Os, CommonCfg.Platform, DEFAULT_MAC_VALUE)
		new := fmt.Sprintf("/%s/%s/%s", CommonCfg.Os, CommonCfg.Platform, CommonCfg.Mac)
		prefix := strings.Replace(key, old, new, 1)
		_, err := EtcdCli.Put(ctx, prefix, value)
		defer cancel()
		if err != nil {
			fmt.Println("配置etcd 失败", err)
			return
		}
	}

}

//	func GetLogCfg()(logcfg LogCfg,err error){
//		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//		//etcd默认配置key:prefix
//		key := fmt.Sprintf("/%s/%s/%s", CommonCfg.Os,CommonCfg.Platform,CommonCfg.Mac)
//		//获取默认配置value
//		resp, err := EtcdCli.Get(ctx, key)
//		defer cancel()
//		if err != nil {
//			fmt.Println("get etcd err:", err)
//			return logcfg,err
//		}
//		logcfg.ClientName = CommonCfg.Mac
//		logcfg.Key = key
//		logcfg.Value = string(resp.Kvs[0].Value)
//		return logcfg,nil
//	}
// func ParseValue(value string) (logkvslice LogKeyValueSlice) {
// 	err := json.Unmarshal([]byte(value), &logkvslice)
// 	if err != nil {
// 		fmt.Println("unmarshal logkeyvalue fail,err:", err)
// 	}
// 	return
// }

func GetLogCfg() (logcfg LogCfg, err error) {
	prefix := fmt.Sprintf("/%s/%s/%s", CommonCfg.Os, CommonCfg.Platform, CommonCfg.Mac)
	logcfg, err = GetLogCfgByPrefix(prefix)
	if err != nil {
		fmt.Println("get logcfg err:", err)
		return
	}
	return
}
