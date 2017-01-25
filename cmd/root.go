// Copyright © 2016 Zhang Peihao <zhangpeihao@gmail.com>

package cmd

import (
	"flag"
	"fmt"
	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
	"github.com/zhangpeihao/zim/pkg/util"
	"net/http"
	"os"

	// Register all serializers
	_ "github.com/zhangpeihao/zim/pkg/protocol/serialize/register"
)

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
	if viper.GetBool("debug") {
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

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default /etc/zim.yaml)")

	RootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose mode")
	viper.BindPFlag("verbose", RootCmd.PersistentFlags().Lookup("verbose"))

	RootCmd.PersistentFlags().BoolP("debug", "d", false, "debug mode. runtime profiling data at: htpp://localhost:8766/debug/pprof")
	viper.BindPFlag("debug", RootCmd.PersistentFlags().Lookup("debug"))

	RootCmd.PersistentFlags().String("vmodule", "", "vmodule for glog")
	viper.BindPFlag("vmodule", RootCmd.PersistentFlags().Lookup("vmodule"))

	RootCmd.PersistentFlags().String("log_dir", "", "log path")
	viper.BindPFlag("log_dir", RootCmd.PersistentFlags().Lookup("log_dir"))

	RootCmd.PersistentFlags().Int("log_level", 3, "log level (0: info, 1: warning, 2: error, 3:fatal)")
	viper.BindPFlag("log_level", RootCmd.PersistentFlags().Lookup("log_level"))

	RootCmd.PersistentFlags().Int("cpu", 1, "the number of logical CPUs used by the current process")
	viper.BindPFlag("cpu", RootCmd.PersistentFlags().Lookup("cpu"))

	viper.AutomaticEnv() // read in environment variables that match

}

var initConfigDone bool

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if initConfigDone {
		return
	}
	initConfigDone = true

	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
		viper.SetConfigType("yaml")
	} else {
		viper.SetConfigName("zim")  // name of config file (without extension)
		viper.AddConfigPath("/etc") // adding home directory as first search path
	}
	viper.ReadInConfig()

	flag.Set("v", viper.GetString("log_level"))

	if viper.GetBool("verbose") {
		jww.SetStdoutThreshold(jww.LevelTrace)
		flag.Set("v", "4")
		flag.Set("alsologtostderr", "true")
	}
	if len(viper.GetString("vmodule")) > 0 {
		flag.Set("vmodule", viper.GetString("vmodule"))
	}
	if len(viper.GetString("log_dir")) > 0 {
		flag.Set("log_dir", viper.GetString("log_dir"))
	}
	util.SetCPU(viper.GetInt("cpu"))
}
