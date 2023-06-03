package main

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	logger "goPandora/internal/log"
)

func main() {
	// 初始化logger
	err := logger.InitLogger("info")
	if err != nil {
		panic(err)
	}

	// 绑定命令行参数
	pflag.StringP("server", "s", "127.0.0.1:8080", "server address")
	pflag.StringSliceP("proxys", "p", nil, "proxy address")
	pflag.String("CHATGPT_API_PREFIX", "https://ai.fakeopen.com", "CHATGPT_API_PREFIX")
	pflag.Parse()

	// 初始化Viperr
	err = viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		logger.Error("viper.BindPFlags failed", zap.Error(err))
		return
	}

	// 读取命令行参数的值
	server := viper.GetString("server")
	proxies := viper.GetStringSlice("proxys")
	gptPre := viper.GetString("CHATGPT_API_PREFIX")

	// 打印结果
	logger.Info("server", zap.String("server", server))
	logger.Info("proxys", zap.Strings("proxys", proxies))
	logger.Info("CHATGPT_API_PREFIX", zap.String("CHATGPT_API_PREFIX", gptPre))
}
