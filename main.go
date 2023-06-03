package main

import (
	"fmt"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func main() {
	// 绑定命令行参数
	pflag.StringP("server", "s", "127.0.0.1:8080", "server address")
	pflag.StringSliceP("proxys", "p", nil, "proxy address")
	pflag.String("CHATGPT_API_PREFIX", "https://ai.fakeopen.com", "CHATGPT_API_PREFIX")
	pflag.Parse()

	// 初始化Viperr
	err := viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		return
	}

	// 读取命令行参数的值
	server := viper.GetString("server")
	proxies := viper.GetStringSlice("proxys")
	gptPre := viper.GetString("CHATGPT_API_PREFIX")

	// 打印结果
	fmt.Println("Server:", server)
	fmt.Println("Proxies:", proxies)
	fmt.Println("CHATGPT API PEFIX:", gptPre)
}
