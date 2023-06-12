package db

import (
	"fmt"
	"github.com/google/uuid"
	"go.uber.org/zap"
	logger "goPandora/internal/log"
	"goPandora/internal/pandora"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"os"
	"strings"
	"time"
)

type SubEnum string

const (
	Google  SubEnum = "google-oauth2"
	Outlook SubEnum = "windowslive"
	OpenAI  SubEnum = "auth0"
)

type User struct {
	ID           uint `gorm:"primary_key:autoIncrement"`
	Email        string
	Password     string
	UserID       string  `gorm:"unique"`
	Sub          SubEnum `gorm:type:enum("google-oauth2","windowslive", "auth0") default:"auth0"`
	Token        string
	RefreshToken string
	UpdatedTime  time.Time `gorm:"autoUpdateTime"`
	ExpiryTime   time.Time
	Comment      string
}

type UserToken struct {
	UUID    uuid.UUID `gorm:"primaryKey;type:char(36);not null;unique"`
	UserID  string
	Token   string
	Comment string
}

var db *gorm.DB

// InitSQLite 初始化SQLite
func InitSQLite(dbFilePath string) error {
	// 判断数据库文件是否存在
	_, err := os.Stat(dbFilePath)
	if os.IsNotExist(err) {
		logger.Info("Creating new database file...", zap.String("dbFilePath", dbFilePath))
		_, err := os.Create(dbFilePath)
		if err != nil {
			return fmt.Errorf("failed to create database file: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to check database file: %w", err)
	}

	// 打开数据库连接
	db, err = gorm.Open(sqlite.Open(dbFilePath), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}
	return nil
}

// GetDB 获取数据库操作指针
func GetDB() (*gorm.DB, error) {
	if nil == db {
		return nil, fmt.Errorf("database connection is not initialized")
	}
	return db, nil
}

// CloseDB 关闭数据库连接
func CloseDB() {
	if nil != db {
		sqlDB, _ := db.DB()
		err := sqlDB.Close()
		if err != nil {
			return
		}
	}
}

// BeforeCreate 向User表插入数据后自动添加UUID
func (u *UserToken) BeforeCreate(tx *gorm.DB) error {
	u.UUID = uuid.New()
	return nil
}

// AddUser 添加用户
func AddUser(refreshToken string, email string, password string, comment string) error {
	var token string
	if refreshToken == "" {
		token, refreshToken, _ = pandora.Auth0(email, password, "", "")
	} else {
		token, _ = pandora.GetTokenByRefreshToken(refreshToken)
	}
	payload, err := pandora.CheckAccessToken(token)
	if err != nil {
		logger.Error("pandora.GetTokenByRefreshToken failed", zap.Error(err))
		return err
	}
	exp, _ := payload["exp"].(float64)
	expires := time.Unix(int64(exp), 0)
	userId := payload["https://api.openai.com/auth"].(map[string]interface{})["user_id"].(string)
	sub := payload["sub"].(string)
	index := strings.Index(sub, "|")
	if index != -1 {
		sub = sub[:index]
	}

	user := &User{
		Email:        email,
		Password:     password,
		UserID:       userId,
		Sub:          SubEnum(sub),
		Token:        token,
		RefreshToken: refreshToken,
		ExpiryTime:   expires,
		Comment:      comment,
	}
	res := db.FirstOrCreate(&user, User{UserID: user.UserID})
	if res.Error != nil {
		logger.Error("sqlite.FirstOrCreate failed", zap.Error(res.Error))
		return res.Error
	}
	if res.RowsAffected > 0 {
		logger.Info("add user success and uuid is", zap.String("user id", user.UserID))
	} else {
		logger.Info("The record already exists and the insert operation is skipped and uuid is", zap.String("user id", user.UserID))
	}
	createUserTokenMap(token, userId, comment)

	return nil
}

func createUserTokenMap(token string, userId string, comment string) {
	userToken := &UserToken{
		Token:   token,
		UserID:  userId,
		Comment: comment,
	}
	userTokenRes := db.FirstOrCreate(&userToken, UserToken{UserID: userId})
	if userTokenRes.Error != nil {
		logger.Error("sqlite.FirstOrCreate failed", zap.Error(userTokenRes.Error))
	}
	if userTokenRes.RowsAffected > 0 {
		logger.Info("add user token success and uuid is", zap.String("user uuid", userToken.UUID.String()))
	} else {
		logger.Info("The record already exists and the insert operation is skipped and uuid is", zap.String("user token id", userToken.UserID))
	}
}

// GetTokenAndExpiryTimeByUUID 根据UUID获取token与过期时间
func GetTokenAndExpiryTimeByUUID(uuid string) (string, time.Time, error) {
	var userToken struct {
		Token      string
		ExpiryTime time.Time
	}

	db.Table("user_tokens").
		Select("user_tokens.token, users.expiry_time").
		Joins("JOIN users ON user_tokens.user_id = users.user_id").
		Where("user_tokens.uuid = ?", uuid).
		First(&userToken)

	return userToken.Token, userToken.ExpiryTime, db.Error
}
