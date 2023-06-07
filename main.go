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
	"gorm.io/gorm"
	"os"
	"strings"
	"time"
)

func main() {
	// 初始化logger
	logger.InitLogger("debug")

	// 绑定命令行参数
	pflag.StringP("server", "s", ":8080", "server address")
	pflag.StringSliceP("proxys", "p", nil, "proxy address")
	pflag.StringP("database", "b", "./data.db", "database file path")
	pflag.String("CHATGPT_API_PREFIX", "https://ai.fakeopen.com", "CHATGPT_API_PREFIX")
	pflag.String("add-user-file", "", "add user file path")
	pflag.Bool("add-user", false, "add user")
	pflag.Bool("list-user", false, "list user")
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

	if viper.GetBool("add-user") { // 添加用户
		email := readerStringByCMD("Email:")
		password := readerStringByCMD("Password:")
		refreshToken := readerStringByCMD("RefreshToken:")
		comment := readerStringByCMD("Comment:") // 备注

		if addUser(refreshToken, email, password, comment, sqlite) == nil {
			return
		}
	} else if viper.GetString("add-user-file") != "" { // 读取文件添加用户
		filePath := viper.GetString("add-user-file")
		// 读取配置文件
		addUserByFile(filePath, sqlite)
		return
	} else if viper.GetBool("list-user") {
		var users []db.User
		sqlite.Find(&users)
		for _, user := range users {
			fmt.Printf("Email: %s, UUID: %s\n", user.Email, user.UUID)
		}
	} else {
		cloudParam := web.PandoraParam{
			ApiPrefix:     gptPre,
			PandoraSentry: "false",
			BuildId:       "cx416mT2Lb0ZTj5FxFg1l",
		}
		web.ServerStart(server, &cloudParam)
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
		_ = addUser(refreshToken, email, password, notes, sqlite)
	}
	if err := scanner.Err(); err != nil {
		logger.Fatal("scanner.Err failed", zap.Error(err))
		return
	}
	return
}

// addUser 添加用户
func addUser(refreshToken string, email string, password string, comment string, sqlite *gorm.DB) error {
	token, _ := pandora.GetTokenByRefreshToken(refreshToken)
	payload, err := pandora.CheckAccessToken(token)
	if err != nil {
		logger.Error("pandora.GetTokenByRefreshToken failed", zap.Error(err))
		return err
	}
	exp, _ := payload["exp"].(float64)
	expires := time.Unix(int64(exp), 0)

	user := &db.User{
		Email:        email,
		Password:     password,
		Token:        token,
		RefreshToken: refreshToken,
		ExpiryTime:   expires,
		Comment:      comment,
	}
	res := sqlite.FirstOrCreate(&user, db.User{Email: user.Email})
	if res.Error != nil {
		logger.Error("sqlite.FirstOrCreate failed", zap.Error(res.Error))
		return res.Error
	}
	if res.RowsAffected > 0 {
		logger.Info("add user success")
	} else {
		logger.Info("The record already exists and the insert operation is skipped")
	}
	return nil
}

func readerStringByCMD(printString string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(printString)
	str, _ := reader.ReadString('\n')
	str = strings.TrimRight(str, "\r\n")
	return str
}
