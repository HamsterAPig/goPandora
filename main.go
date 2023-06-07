package main

import (
	"bufio"
	"fmt"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"goPandora/internal/db"
	logger "goPandora/internal/log"
	"goPandora/internal/pandora"
	"goPandora/web"
	"os"
	"strings"
)

func main() {
	// 初始化logger
	logger.InitLogger("debug")

	// 绑定命令行参数
	pflag.StringP("server", "s", ":8080", "server address")
	pflag.StringSliceP("proxys", "p", nil, "proxy address")
	pflag.StringP("database", "b", "./data.db", "database file path")
	pflag.String("CHATGPT_API_PREFIX", "https://ai.fakeopen.com", "CHATGPT_API_PREFIX")
	pflag.Bool("add-user", false, "add user")
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
	dbFilePath := viper.GetString("database")

	// 打印结果
	logger.Debug("server", zap.String("server", server))
	logger.Debug("proxys", zap.Strings("proxys", proxies))
	logger.Debug("database", zap.String("database", dbFilePath))
	logger.Debug("CHATGPT_API_PREFIX", zap.String("CHATGPT_API_PREFIX", gptPre))

	err = db.InitSQLite(dbFilePath)
	if err != nil {
		logger.Error("db.InitSQLite failed", zap.Error(err))
		return
	}
	defer db.CloseDB()
	sqlite, _ := db.GetDB()
	err = sqlite.AutoMigrate(&db.User{})
	if err != nil {
		logger.Error("sqlite.AutoMigrate failed", zap.Error(err))
		return
	}

	if viper.GetBool("add-user") {
		email := readerStringByCMD("Email:")
		password := readerStringByCMD("Password:")
		refreshToken := readerStringByCMD("RefreshToken:")

		token, _ := pandora.GetTokenByRefreshToken(refreshToken)
		payload, err := pandora.CheckAccessToken(token)
		if err != nil {
			logger.Error("pandora.GetTokenByRefreshToken failed", zap.Error(err))
			return
		}
		exp, _ := payload["exp"].(int64)

		user := &db.User{
			Email:        email,
			Password:     password,
			Token:        token,
			RefreshToken: refreshToken,
			ExpiryTime:   exp,
		}
		sqlite.Create(&user)
	} else {
		cloudParam := web.PandoraParam{
			ApiPrefix:     gptPre,
			PandoraSentry: "false",
			BuildId:       "cx416mT2Lb0ZTj5FxFg1l",
		}
		web.ServerStart(server, &cloudParam)
	}
}

func readerStringByCMD(printString string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(printString)
	str, _ := reader.ReadString('\n')
	str = strings.TrimRight(str, "\r\n")
	return str
}
