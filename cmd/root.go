/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"liuyifei/pkg/config"
	"liuyifei/pkg/util"
	"log"
	"os"

	"github.com/hpcloud/tail"
	"github.com/spf13/cobra"
	etcdClient "go.etcd.io/etcd/clientv3"
)

const ()

var (
	cfgfile string

	endpoints   []string
	dialtimeout int64
	platform    string
	sendto      string
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
	rootCmd.PersistentFlags().StringVarP(&cfgfile, "config", "c", "", "config file of logagent") //d:/go/liuyifei/liuyifei.ini
	rootCmd.PersistentFlags().StringVarP(&platform, "platform", "p", "system", "platform name")
	rootCmd.PersistentFlags().StringSliceVarP(&endpoints, "etcd", "e", []string{"172.16.1.200:2379"}, "etcd endpoints eg 172.16.1.200:2379")
	rootCmd.PersistentFlags().StringVarP(&sendto, "sendto", "", "kafka", "logs sendto who")
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
		//初始化配置文件

		cfg, err := getCommonCfg()
		if err != nil {
			log.Fatalf("获取配置文件失败!err:%v\n", err)
		}

		//从ETCD获取日志监控项
		etcdcli, err := config.GetClientEtcd(cfg)
		defer etcdcli.Close()
		//如果没有获取到该设备mac下的日志配置，就用默认配置设置一次。
		prefix := fmt.Sprintf("/%s/%s/%s", cfg.Os, cfg.Platform, cfg.Mac)
		defaullogcfg, err := config.GetLogCfgByPrefix(prefix)
		if len(defaullogcfg.LogPath) == 0 {
			config.SetDefaultLogCfg()
		}
		//获取该设备下的日志配置
		logcfg, err := config.GetLogCfg()
		if err != nil {
			return
		}

		//初始化contextMap,跟踪每个tail任务
		ctxmap := make(map[string]config.CtxS)

		//初始化kafkaclient
		kafkaclient, err := config.GetClientKafka(cfg.Address)
		//defer kafkaclient.Close()
		if err != nil {
			log.Fatalf("连接kafka失败%v\n", err)
		}
		kafkaproducer := config.KafkaProducer{
			Client: kafkaclient,
		}

		serviceCfg := &ServiceConfig{}
		serviceCfg.Cfg = cfg
		serviceCfg.EtcdCli = etcdcli
		serviceCfg.CtxMap = ctxmap
		serviceCfg.KafkaProducer = &kafkaproducer
		serviceCfg.LogCfg = logcfg
		fmt.Println("serviceCfg: ",serviceCfg)
		StartSendMessage(serviceCfg)
		WatchEtcd(serviceCfg)
		// for key, filepath := range logcfg.LogPath {
		// 	//fmt.Println(key, filepath)
		// 	ctx, cancel := context.WithCancel(context.Background())
		// 	ctxmap[key] = config.CtxS{
		// 		Ctx:    ctx,
		// 		Cancel: cancel,
		// 	}
		// 	lineCh, err := config.ReadLogByTail(filepath)
		// 	if err != nil {
		// 		fmt.Printf("tail 失败err:%v", err)
		// 		close(lineCh)
		// 		continue
		// 	}
		// 	go kafkaproducer.SendToKafka(ctx, cfg.Platform, lineCh)
		// }

		// watchChan := etcdcli.Watch(context.Background(), fmt.Sprintf("/%s", cfg.Platform), etcdClient.WithPrefix())
		// for {
		// 	select {
		// 	case resp := <-watchChan:
		// 		for _, event := range resp.Events {
		// 			fi, err := os.Stat(string(event.Kv.Value))
		// 			if err != nil {

		// 			} else {
		// 				if fi.IsDir() {
		// 					fmt.Printf("这是一个目录，请重新配置\n")
		// 					continue
		// 				}
		// 			}
		// 			//取消之前的任务
		// 			if ctx, ok := ctxmap[string(event.Kv.Key)]; ok {
		// 				ctx.Cancel()
		// 				fmt.Printf("任务已取消%s%s\n", event.Kv.Key, event.Kv.Value)

		// 			}
		// 			logcfg.LogPath[string(event.Kv.Key)] = string(event.Kv.Value)
		// 			//根据最新获取的etcd配置，执行任务
		// 			ctx, cancel := context.WithCancel(context.Background())
		// 			ctxmap[string(event.Kv.Key)] = config.CtxS{
		// 				Ctx:    ctx,
		// 				Cancel: cancel,
		// 			}

		// 			lineCh, err := config.ReadLogByTail(string(event.Kv.Value))
		// 			if err != nil {
		// 				fmt.Println("ReadLogByTail err:", err)
		// 				close(lineCh)
		// 				continue
		// 			}
		// 			go kafkaproducer.SendToKafka(ctx, cfg.Platform, lineCh)
		// 		}
		// 	}
		// }

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
func StartSendMessage(servicecfg *ServiceConfig) {
	for key, filepath := range servicecfg.LogCfg.LogPath {
		//fmt.Println(key, filepath)
		ctx, cancel := context.WithCancel(context.Background())
		servicecfg.CtxMap[key] = config.CtxS{
			Ctx:    ctx,
			Cancel: cancel,
		}
		lineCh, err := config.ReadLogByTail(filepath)
		if err != nil {
			fmt.Printf("tail 失败err:%v", err)
			close(lineCh)
			continue
		}

		SentMessage(ctx, servicecfg, lineCh)
	}
}
func SentMessage(ctx context.Context, serviceCfg *ServiceConfig, lineCh chan *tail.Line) {
	switch serviceCfg.Cfg.Sendto {
	case "kafka":
		fmt.Printf("发送日志到kafka\n")
		go serviceCfg.KafkaProducer.SendToKafka(ctx, serviceCfg.Cfg.Platform, lineCh)
	case "mysql":
		fmt.Printf("发送日志到mysql功能暂未实现\n")
	default:
		//go serviceCfg.KafkaProducer.SendToKafka(ctx, serviceCfg.Cfg.Platform, lineCh)
	}
}

func WatchEtcd(servicecfg *ServiceConfig) {
	prefix := fmt.Sprintf("/%s/%s/%s", servicecfg.Cfg.Os, servicecfg.Cfg.Platform, servicecfg.Cfg.Mac)
	fmt.Println("watch prefix:",prefix)
	watchChan := servicecfg.EtcdCli.Watch(context.Background(), prefix, etcdClient.WithPrefix())
	for {
		select {
		case resp := <-watchChan:
			for _, event := range resp.Events {
				// fi, err := os.Stat(string(event.Kv.Value))
				// if err != nil {

				// } else {
				// 	if fi.IsDir() {
				// 		fmt.Printf("这是一个目录，请重新配置\n")
				// 		continue
				// 	}
				// }
				if util.IsDir(string(event.Kv.Value)) {
					continue
				}
				//删除之前的任务
				if ctx, ok := servicecfg.CtxMap[string(event.Kv.Key)]; ok {
					fmt.Println("删除之前的任务", ctx.Cancel)
					ctx.Cancel()
					fmt.Printf("任务已取消%s%s\n", event.Kv.Key, event.Kv.Value)
					delete(servicecfg.CtxMap, string(event.Kv.Key))

				}
				if event.Kv.Value == nil {
					delete(servicecfg.LogCfg.LogPath, string(event.Kv.Key))
					continue
				}
				servicecfg.LogCfg.LogPath[string(event.Kv.Key)] = string(event.Kv.Value)
				//根据最新获取的etcd配置，执行任务
				newctx, newcancel := context.WithCancel(context.Background())
				servicecfg.CtxMap[string(event.Kv.Key)] = config.CtxS{
					Ctx:    newctx,
					Cancel: newcancel,
				}

				newlineCh, err := config.ReadLogByTail(string(event.Kv.Value))
				fmt.Println("通道状态", newlineCh)
				if err != nil {
					fmt.Println("ReadLogByTail err:", err)
					close(newlineCh)
					continue
				}
				fmt.Println(servicecfg.CtxMap, servicecfg.LogCfg.LogPath)
				SentMessage(newctx, servicecfg, newlineCh)
			}
		}
	}
}

func getCommonCfgFromCmd() (cfg config.CommonConfig, err error) {
	cfg = config.GetDefaultCommonConfig()
	cfg.Endpoints = endpoints
	//cfg.DialTimeout = dialtimeout
	cfg.Platform = platform
	cfg.Sendto = sendto
	cfg.Address = address
	return
}

func getCommonCfg() (cfg config.CommonConfig, err error) {
	cfg = config.CommonConfig{}

	if cfgfile != "" {
		cfg, err = config.GetCommonCfgFromIni(cfgfile)
		fmt.Printf("get config from file.ini\n")
		if err != nil {
			fmt.Println(err)
			return
		}
	} else {
		cfg, err = getCommonCfgFromCmd()
		fmt.Printf("get config from cmd\n")
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	cfg.SystemInfo = config.GetSystemInfo(cfg.Endpoints)
	config.CommonCfg = cfg
	return
}
