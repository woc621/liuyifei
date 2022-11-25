package config

import (
	"context"
	"fmt"

	"github.com/go-ini/ini"
	//"json/encoding"
)

type CommonConfig struct {
	//etcd
	Endpoints   []string `ini:"endpoints" json:"endpoints"`
	DialTimeout int64    `ini:"dialtimeout" json:"dialtimeout"`
	//p
	Platform string `ini:"platform" json:"platform"`
	//kafka
	Address []string `ini:"address" json:"address"`
}
type CtxS struct {
	Ctx    context.Context
	Cancel context.CancelFunc
}

func GetDefaultCommonConfig() CommonConfig {
	return CommonConfig{
		Endpoints:   []string{"172.16.1.200:2379"},
		DialTimeout: 5,
		Platform:    "system",
		Address:     []string{"172.16.1.200:9092"},
	}
}

func GetCommonCfgFromIni(source interface{}) (CommonConfig, error) {
	cfg := GetDefaultCommonConfig()
	fmt.Println("默认配置：", cfg)
	err := ini.MapTo(&cfg, source)
	if err != nil {
		return CommonConfig{}, err
	}
	fmt.Println("从ini文件获取的配置:", cfg)
	return cfg, nil
}
