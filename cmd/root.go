// Copyright © 2016 Zhang Peihao <zhangpeihao@gmail.com>

package cmd

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
)

var cfgFile string
var cfgVerbose bool
var cfgVmodule, cfgLogDir string
var cfgDebug bool

// RootCmd root命令
var RootCmd = &cobra.Command{
	Use:   "zim",
	Short: "IM服务",
	Long: `IM集群服务

包括一些模块：
gateway：网关。提供TCP, UDP和WebSocket等接入方式，与客户端
         建立稳定的双向连接。
maintain：网控。实时监控集群各个服务的状态`,
}

// Execute 执行命令
func Execute() {
	if cfgDebug {
		go func() {
			fmt.Println(http.ListenAndServe("localhost:8870", nil))
		}()
	}
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.zim.yaml)")
	RootCmd.PersistentFlags().BoolVarP(&cfgVerbose, "verbose", "v", false, "verbose mode")
	RootCmd.PersistentFlags().BoolVarP(&cfgDebug, "debug", "d", false, "debug mode. runtime profiling data at: htpp://localhost:8766/debug/pprof")
	RootCmd.PersistentFlags().StringVar(&cfgVmodule, "vmodule", "", "vmodule for glog")
	RootCmd.PersistentFlags().StringVar(&cfgLogDir, "log_dir", "", "log path")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
		viper.SetConfigType("yaml")
	} else {
		viper.SetConfigName(".zim")  // name of config file (without extension)
		viper.AddConfigPath("$HOME") // adding home directory as first search path
	}
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
		if viper.InConfig("debug") {
			cfgDebug = viper.GetBool("debug")
		}
		if viper.InConfig("verbose") {
			cfgVerbose = viper.GetBool("verbose")
		}
		if viper.InConfig("router-json") {
			cfgRouterJSON = viper.GetString("router-json")
		}
		if viper.InConfig("wss-cert-file") {
			cfgCertFile = viper.GetString("wss-cert-file")
		}
		if viper.InConfig("wss-key-file") {
			cfgKeyFile = viper.GetString("wss-key-file")
		}

	} else {
		fmt.Println("ReadInConfig error:", err)
		os.Exit(1)
	}

	if cfgVerbose {
		jww.SetStdoutThreshold(jww.LevelTrace)
		flag.Set("v", "4")
		flag.Set("alsologtostderr", "true")
	}
	if len(cfgVmodule) > 0 {
		flag.Set("vmodule", cfgVmodule)
	}
	if len(cfgLogDir) > 0 {
		flag.Set("log_dir", cfgLogDir)
	}
}
