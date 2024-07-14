package main

import (
	"context"
	"github.com/dadaxiaoxiao/tag/ioc"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"net/http"
	"time"
)

func main() {
	initViper()
	initPrometheus()
	closeFunc := ioc.InitOTEL()
	app := InitApp()
	err := app.GRPCServer.Serve()
	if err != nil {
		panic(err)
	}

	// 下面这些是正常退出
	// 一分钟内要关完，且退出
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	closeFunc(ctx)
}

func initViper() {
	cfile := pflag.String("config", "config/config.yaml", "配置文件路径")
	pflag.Parse()
	// 直接指定文件路径
	viper.SetConfigFile(*cfile)
	// 实时监听配置变更
	viper.WatchConfig()
	// 读取配置到viper 里面
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}

func initViperRemote() {
	type Config struct {
		Provider string `yaml:"provider"`
		Endpoint string `yaml:"endpoint"`
		Path     string `yaml:"path"`
	}

	var config Config
	err := viper.UnmarshalKey("remoteProvider", &config)
	if err != nil {
		panic(err)
	}
	// 新增远程配置
	err = viper.AddRemoteProvider(config.Provider,
		config.Endpoint, config.Path)
	if err != nil {
		panic(err)
	}
	viper.SetConfigType("yaml")
	// 实时监听配置变更
	err = viper.WatchRemoteConfig()
	if err != nil {
		panic(err)
	}
	// 读取配置到viper 里面
	err = viper.ReadRemoteConfig()
	if err != nil {
		panic(err)
	}
}

func initPrometheus() {
	type Config struct {
		ListenPort string `yaml:"listenPort"`
	}
	var config Config
	err := viper.UnmarshalKey("prometheus", &config)
	if err != nil {
		panic(err)
	}
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		// 暴露监听端口
		http.ListenAndServe(config.ListenPort, nil)
	}()
}
