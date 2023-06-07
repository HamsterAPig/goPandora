package db

import (
	"fmt"
	"github.com/google/uuid"
	"go.uber.org/zap"
	logger "goPandora/internal/log"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"os"
	"time"
)

type User struct {
	ID           uint `gorm:"primary_key:autoIncrement"`
	Email        string
	Password     string
	UUID         string `gorm:"unique"`
	Token        string
	RefreshToken string
	UpdatedTime  time.Time `gorm:"autoUpdateTime"`
	ExpiryTime   time.Time
	Comment      string
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
func (u *User) BeforeCreate(tx *gorm.DB) error {
	u.UUID = uuid.New().String()
	return nil
}
