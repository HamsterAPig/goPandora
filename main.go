package main

import (
	"fmt"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	logger "goPandora/internal/log"
	"goPandora/web"
)

func main() {
	// 初始化logger
	logger.InitLogger("debug")

	// 绑定命令行参数
	pflag.StringP("server", "s", "127.0.0.1:8080", "server address")
	pflag.StringSliceP("proxys", "p", nil, "proxy address")
	pflag.String("CHATGPT_API_PREFIX", "https://ai.fakeopen.com", "CHATGPT_API_PREFIX")
	pflag.Parse()

	// 初始化Viperr
	err := viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		logger.Error("viper.BindPFlags failed", zap.Error(err))
		return
	}

	// 读取命令行参数的值
	server := viper.GetString("server")
	proxies := viper.GetStringSlice("proxys")
	gptPre := viper.GetString("CHATGPT_API_PREFIX")

	// 打印结果
	logger.Debug("server", zap.String("server", server))
	logger.Debug("proxys", zap.Strings("proxys", proxies))
	logger.Debug("CHATGPT_API_PREFIX", zap.String("CHATGPT_API_PREFIX", gptPre))

	//cloudParam := web.PandoraParam{
	//	ApiPrefix:     gptPre,
	//	PandoraSentry: false,
	//	BuildId:       "",
	//}
	token, err := web.CheckAccessToken("eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6Ik1UaEVOVUpHTkVNMVFURTRNMEZCTWpkQ05UZzVNRFUxUlRVd1FVSkRNRU13UmtGRVFrRXpSZyJ9.eyJodHRwczovL2FwaS5vcGVuYWkuY29tL3Byb2ZpbGUiOnsiZW1haWwiOiJqZGNtdnJ0Yjc0cmN0OHRxc29keTZtYnNwN0Bwcm90b24ubWUiLCJlbWFpbF92ZXJpZmllZCI6dHJ1ZX0sImh0dHBzOi8vYXBpLm9wZW5haS5jb20vYXV0aCI6eyJ1c2VyX2lkIjoidXNlci1IczFpckQxSXd5WFE3bTNCTTlWcmxSQjAifSwiaXNzIjoiaHR0cHM6Ly9hdXRoMC5vcGVuYWkuY29tLyIsInN1YiI6ImF1dGgwfDY0Mzg4NjJjYzQwYTEyZDQ0NDRhNzQyMiIsImF1ZCI6WyJodHRwczovL2FwaS5vcGVuYWkuY29tL3YxIiwiaHR0cHM6Ly9vcGVuYWkub3BlbmFpLmF1dGgwYXBwLmNvbS91c2VyaW5mbyJdLCJpYXQiOjE2ODU1ODEyMjQsImV4cCI6MTY4Njc5MDgyNCwiYXpwIjoicGRsTElYMlk3Mk1JbDJyaExoVEU5VlY5Yk45MDVrQmgiLCJzY29wZSI6Im9wZW5pZCBwcm9maWxlIGVtYWlsIG1vZGVsLnJlYWQgbW9kZWwucmVxdWVzdCBvcmdhbml6YXRpb24ucmVhZCBvZmZsaW5lX2FjY2VzcyJ9.RqEDFX57G0VOmclT1fPzVuc6Wwmq4cBMUbMTFiK_BF5jqMv5Bzs72Aict8YesR4ZU1_RBmQ9YXs4TBA2-ZQrR8LsZ_3vtayGzXMCBbH47f65y9VcE7OjjBulrr5g4EZv_exw-UpeSl9S1eRWyjxrY1RV4icZtWk0dxTQeVYGYyT4Y1fcboXNE-ekxgCkK-Rhd_eR_B6mN2Y2xrJ8E7ytNJApDooaOULzuRR5gKvvWHMg7qe81r0MJdGaW5eXcXabfp_JuYipV8BqVcYXmCvo6ybtpkoZ_pLmJrGVKp_FxaiCet2LmYtJHZ2eJ7ll2DlKKS5Fy81DlJ02kXqslCs2rw")
	if err != nil {
		return
	}
	fmt.Println(token)
	//web.ServerStart(server, &cloudParam)
}
