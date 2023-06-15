package main

import (
	"bufio"
	"fmt"
	"go.uber.org/zap"
	"goPandora/config"
	"goPandora/internal/db"
	logger "goPandora/internal/log"
	"goPandora/web"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	var err error
	config.Conf, err = config.ReadConfig()
	if err != nil {
		if err.Error() == "no enough arguments" {
			return
		} else {
			fmt.Println("read config failed")
			panic(err)
			return
		}
	}

	logger.InitLogger(config.Conf.MainConfig.DebugLevel)
	err = db.InitSQLite(config.Conf.MainConfig.DatabasePath)
	if err != nil {
		logger.Error("db.InitSQLite failed", zap.Error(err))
		return
	}
	defer db.CloseDB()
	sqlite, _ := db.GetDB()
	err = sqlite.AutoMigrate(&db.User{}, &db.UserToken{}, &db.ShareToken{})
	if err != nil {
		logger.Error("sqlite.AutoMigrate failed", zap.Error(err))
		return
	}
	if config.Conf.MainConfig.UserAdd { // 添加用户
		email := readerStringByCMD("Email:")
		password := readerStringByCMD("Password:")
		refreshToken := readerStringByCMD("RefreshToken:")
		comment := readerStringByCMD("Comment:") // 备注

		if db.AddUser(refreshToken, email, password, comment) != nil {
			logger.Error("db.AddUser failed", zap.Error(err))
			return
		}
	} else if config.Conf.MainConfig.UserAddByFilePath != "" { // 读取文件添加用户
		filePath := config.Conf.MainConfig.UserAddByFilePath
		// 读取配置文件
		addUserByFile(filePath)
		return
	} else if config.Conf.MainConfig.ShareTokenAddByID {
		ids := readerStringByCMD("Enter ShareToken ID(s):")
		// 拆分字符串为起始和结束值
		rangeValues := strings.Split(ids, "-")
		start, _ := strconv.Atoi(rangeValues[0])
		end, _ := strconv.Atoi(rangeValues[1])

		// 构建数字切片
		var numbers []int
		for i := start; i <= end; i++ {
			numbers = append(numbers, i)
		}
		uniqueName := readerStringByCMD("Unique Name:")
		siteLimit := readerStringByCMD("Site Limit:")
		ExpireTimeStr := readerStringByCMD("Expire Time:")
		ExpireTime, err := strconv.Atoi(ExpireTimeStr)
		if err != nil {
			logger.Fatal("strconv.Atoi failed", zap.Error(err))
		}
		comment := readerStringByCMD("Comment:")
		logger.Debug("Enter param",
			zap.Ints("Numbers", numbers),
			zap.String("uniqueName", uniqueName),
			zap.String("siteLimit", siteLimit),
			zap.Int("ExpireTime", ExpireTime),
			zap.String("comment", comment),
		)

		for _, id := range numbers {
			userID, err := db.GetUserIDByID(id)
			if err != nil {
				logger.Error("db.GetUserIDByID failed", zap.Error(err))
			}
			err = db.CreateShareToken(userID, uniqueName, time.Duration(ExpireTime)*time.Second, siteLimit, comment)
		}
	} else if config.Conf.MainConfig.UserList {
		ret := db.ListAllUser()
		for _, item := range ret {
			println(item)
		}
	} else {
		web.ServerStart()
	}
}

// addUserByFile 通过文件添加用户
func addUserByFile(filePath string) {
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
		} else if len(fields) == 2 {
			email = fields[0]
			password = fields[1]
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
