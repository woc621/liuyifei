/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"liuyifei/pkg/config"
	"os"

	"github.com/spf13/cobra"
)

var (
	cfgfile  string
	platform string
)

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
		if cfgfile == "" {
			fmt.Printf("cfgfile is nil\n")

		}
		ectdcli, err := config.InitEtcd(cfgfile)
		if err != nil {
			fmt.Println(err)
			return
		}
		platform := config.GetPlatFormCfg(cfgfile)
		logcfg := config.LogCfg{
			PlatForm: platform,
			LogPath:  make(map[string]string),
		}
		logcfg.GetLogCfgFromEtcd(ectdcli)
		fmt.Println(logcfg)
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

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.liuyifei.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.PersistentFlags().StringVarP(&cfgfile, "config", "c", "", "config file of frps")
	rootCmd.PersistentFlags().StringVarP(&platform, "platform", "p", "system", "config file of frps")
}
