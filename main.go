package main

import (
	"bufio"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"goPandora/internal/db"
	logger "goPandora/internal/log"
	"goPandora/web"
	"gorm.io/gorm"
	"os"
	"strings"
)

func main() {

	// 绑定命令行参数
	pflag.StringP("server", "s", ":8080", "server address")
	pflag.StringSliceP("proxys", "p", nil, "proxy address")
	pflag.StringP("database", "b", "./data.db", "database file path")
	pflag.String("CHATGPT_API_PREFIX", "https://ai.fakeopen.com", "CHATGPT_API_PREFIX")
	pflag.String("user-add-file", "", "add user file path")
	pflag.String("web-user-list", "", "user list file path")
	pflag.String("debug-level", "info", "debug level")
	pflag.Bool("user-add", false, "add user")
	pflag.Bool("user-list", false, "list user")
	pflag.Bool("enable_share_page_verify", true, "enable share page verify")
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

	// 设置gin日志等级
	if viper.GetString("debug-level") == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// 初始化logger
	logger.InitLogger(viper.GetString("debug-level"))
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
	err = sqlite.AutoMigrate(&db.User{}, &db.UserToken{})
	if err != nil {
		logger.Error("sqlite.AutoMigrate failed", zap.Error(err))
		return
	}

	if viper.GetBool("user-add") { // 添加用户
		email := readerStringByCMD("Email:")
		password := readerStringByCMD("Password:")
		refreshToken := readerStringByCMD("RefreshToken:")
		comment := readerStringByCMD("Comment:") // 备注

		if db.AddUser(refreshToken, email, password, comment) != nil {
			logger.Error("db.AddUser failed", zap.Error(err))
			return
		}
	} else if viper.GetString("user-add-file") != "" { // 读取文件添加用户
		filePath := viper.GetString("user-add-file")
		// 读取配置文件
		addUserByFile(filePath, sqlite)
		return
	} else if viper.GetBool("user-list") {
		db.ListAllUser()
	} else {
		web.Param = web.PandoraParam{
			ApiPrefix:     gptPre,
			PandoraSentry: false,
			BuildId:       "cx416mT2Lb0ZTj5FxFg1l",
		}
		// 设置是否启用分享页查看验证
		web.Param.EnableSharePageVerify = viper.GetBool("enable_share_page_verify")
		web.ServerStart(server)
	}
}

// addUserByFile 通过文件添加用户
func addUserByFile(filePath string, sqlite *gorm.DB) {
	file, err := os.Open(filePath)
	if err != nil {
		logger.Fatal("os.Open failed", zap.Error(err))
		return
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	// 创建一个scanner用于逐行读取配置文件
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()         // 循环读取每一行
		fields := strings.Fields(line) // 按空格分割每行字段
		var email, password, refreshToken, notes string
		if len(fields) == 3 {
			email = fields[0]
			password = fields[1]
			refreshToken = fields[2]
			notes = ""
		} else if len(fields) == 4 {
			email = fields[0]
			password = fields[1]
			refreshToken = fields[2]
			notes = fields[3]
		}
		err = db.AddUser(refreshToken, email, password, notes)
		if err != nil {
			logger.Error("db.AddUser failed", zap.Error(err))
			return
		}
	}
	if err := scanner.Err(); err != nil {
		logger.Fatal("scanner.Err failed", zap.Error(err))
		return
	}
	return
}

func readerStringByCMD(printString string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(printString)
	str, _ := reader.ReadString('\n')
	str = strings.TrimRight(str, "\r\n")
	return str
}
