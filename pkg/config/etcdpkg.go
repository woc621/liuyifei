package config

import (
	"context"
	"fmt"
	"liuyifei/pkg/util"
	"time"

	"github.com/go-ini/ini"
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

var (
// cfgfile string
)

type EtcdCfg struct {
	EndPoints   []string      `ini:"endpoints"`
	DialTimeout time.Duration `ini:"dialtimeout"`
}
type LogCfg struct {
	PlatForm string
	LogPath  map[string]string
}

func GetPlatFormCfgFromIni(source interface{}) (platform string, err error) {
	f, err := ini.LoadSources(ini.LoadOptions{
		Insensitive:         false,
		InsensitiveSections: false,
		InsensitiveKeys:     false,
		IgnoreInlineComment: true,
		AllowBooleanKeys:    true,
	}, source)
	if err != nil {
		return platform, err
	}
	s, err := f.GetSection(INI_CFG_PLATFORM)

	if err != nil {
		return platform, err
	}
	key, err := s.GetKey(INI_CFG_PLATFORM_key)
	if err != nil {
		return platform, err
	}
	platform = key.Value()
	return platform, nil
}

func GetPlatFormCfg(cfgfile string) (platform string) {
	platform, err := GetPlatFormCfgFromIni(cfgfile)
	if err != nil {
		fmt.Printf("getplatformfrom file.ini err:%v\n", err)
	}
	if platform == "" {
		fmt.Printf("使用默认配置/system \n")
		platform = DEFAULT_PLATFORM_VALUE
	}
	return
}

func getDefaultEtcdCfg() EtcdCfg {
	return EtcdCfg{
		EndPoints:   []string{"172.16.1.200:2379"},
		DialTimeout: 5 * time.Second,
	}
}

func getEtcdCfgFromIni(source interface{}) (EtcdCfg, error) {
	f, err := ini.LoadSources(ini.LoadOptions{
		Insensitive:         false,
		InsensitiveSections: false,
		InsensitiveKeys:     false,
		IgnoreInlineComment: true,
		AllowBooleanKeys:    true,
	}, source)
	if err != nil {
		return EtcdCfg{}, err
	}
	s, err := f.GetSection("etcd")
	if err != nil {
		return EtcdCfg{}, err
	}
	etcdcfg := getDefaultEtcdCfg()
	err = s.MapTo(&etcdcfg)
	if err != nil {
		return EtcdCfg{}, err
	}
	etcdcfg.DialTimeout *= dialtimeoutunits //转换成秒
	return etcdcfg, nil
}

func initEtcdConfig(etcdcfg EtcdCfg) (config etcdClient.Config) {
	return etcdClient.Config{
		Endpoints:   etcdcfg.EndPoints,
		DialTimeout: etcdcfg.DialTimeout,
	}
}

func InitEtcd(cfgfile string) (etcdCli *etcdClient.Client, err error) {
	var cfg EtcdCfg
	if cfgfile == "" {
		cfg = getDefaultEtcdCfg()

	} else {
		cfg, err = getEtcdCfgFromIni(cfgfile)
		if err != nil {
			fmt.Printf("获取etcd配置文件失败:err%v\n", err)
			return etcdCli, err
		}
	}
	etccfg := initEtcdConfig(cfg)
	etcdCli, err = etcdClient.New(etccfg)
	if err != nil {
		fmt.Println("connect fail err:", err)
		return etcdCli, err
	}
	return etcdCli, err
}

func (logcfg LogCfg) GetLogCfgFromEtcd(etcdCli *etcdClient.Client) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	prefix := fmt.Sprintf("/%s", logcfg.PlatForm)
	resp, err := etcdCli.Get(ctx, prefix, etcdClient.WithPrefix())
	defer cancel()
	if err != nil {
		fmt.Println("get etcd err:", err)
		return
	}
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
}
