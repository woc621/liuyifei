/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"liuyifei/pkg/config"
	"log"
	"os"

	"github.com/spf13/cobra"
	etcdClient "go.etcd.io/etcd/clientv3"
)

const ()

var (
	cfgfile string

	endpoints   []string
	dialtimeout int64
	platform    string
	address     []string
)

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.liuyifei.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.PersistentFlags().StringVarP(&cfgfile, "config", "c", "d:/go/liuyifei/liuyifei.ini", "config file of logagent")
	rootCmd.PersistentFlags().StringVarP(&platform, "platform", "p", "system", "platform name")
	rootCmd.PersistentFlags().StringSliceVarP(&endpoints, "etcd", "e", []string{"172.16.1.200:2379"}, "etcd endpoints eg 172.16.1.200:2379")
	//rootCmd.PersistentFlags().Int64VarP(&dialtimeout, "dialtimeout", "et", 5, "ETCD timeout /s")
	rootCmd.PersistentFlags().StringSliceVarP(&address, "address", "a", []string{"172.16.1.200:9092"}, "kafka address eg 192.168.1.1:9092")
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "liuyifei",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
	Run: func(cmd *cobra.Command, args []string) {
		//获取日志客户端配置文件
		cfg, err := getCommonCfg()
		if err != nil {
			log.Fatalf("获取配置文件失败!err:%v\n", err)
		}
		//从ETCD获取日志监控项
		etcdcli, err := config.GetClientEtcd(cfg)
		defer etcdcli.Close()
		logcfg, err := cfg.GetLogCfg(etcdcli)

		//初始化contextMap,跟踪每个tail任务
		ctxmap := make(map[string]config.CtxS)

		//初始化kafkaclient
		kafkaclient, err := config.GetClientKafka(cfg.Address)
		defer kafkaclient.Close()
		if err != nil {
			log.Fatalf("连接kafka失败%v\n", err)
		}
		kafkaproducer := config.KafkaProducer{
			Client: kafkaclient,
		}

		for key, filepath := range logcfg.LogPath {
			//fmt.Println(key, filepath)
			ctx, cancel := context.WithCancel(context.Background())
			ctxmap[key] = config.CtxS{
				Ctx:    ctx,
				Cancel: cancel,
			}
			lineCh, err := config.ReadLogByTail(filepath)
			if err != nil {
				fmt.Printf("tail 失败err:%v", err)
				close(lineCh)
				continue
			}
			go kafkaproducer.SendToKafka(ctx, cfg.Platform, lineCh)
		}

		watchChan := etcdcli.Watch(context.Background(), fmt.Sprintf("/%s", cfg.Platform), etcdClient.WithPrefix())
		for {
			select {
			case resp := <-watchChan:
				for _, event := range resp.Events {
					fi, err := os.Stat(string(event.Kv.Value))
					if err != nil {

					} else {
						if fi.IsDir() {
							fmt.Printf("这是一个目录，请重新配置\n")
							continue
						}
					}
					//取消之前的任务
					if ctx, ok := ctxmap[string(event.Kv.Key)]; ok {
						ctx.Cancel()
						fmt.Printf("任务已取消%s%s\n", event.Kv.Key, event.Kv.Value)

					}
					logcfg.LogPath[string(event.Kv.Key)] = string(event.Kv.Value)
					//根据最新获取的etcd配置，执行任务
					ctx, cancel := context.WithCancel(context.Background())
					ctxmap[string(event.Kv.Key)] = config.CtxS{
						Ctx:    ctx,
						Cancel: cancel,
					}

					lineCh, err := config.ReadLogByTail(string(event.Kv.Value))
					if err != nil {
						fmt.Println("ReadLogByTail err:", err)
						close(lineCh)
						continue
					}
					go kafkaproducer.SendToKafka(ctx, cfg.Platform, lineCh)
				}
			}
		}

	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func getCommonCfgFromCmd() (cfg config.CommonConfig, err error) {
	cfg = config.GetDefaultCommonConfig()
	cfg.Endpoints = endpoints
	//cfg.DialTimeout = dialtimeout
	cfg.Platform = platform
	cfg.Address = address
	return
}

func getCommonCfg() (cfg config.CommonConfig, err error) {
	cfg = config.CommonConfig{}
	if cfgfile != "" {
		cfg, err = config.GetCommonCfgFromIni(cfgfile)
		if err != nil {
			fmt.Println(err)
			return
		}
	} else {
		cfg, err = getCommonCfgFromCmd()
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	return
}
