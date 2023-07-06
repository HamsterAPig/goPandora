package config

import (
	"fmt"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type VariableConfig struct {
	MainConfig       MainSection       `mapstructure:"main"`
	CloudflareConfig CloudflareSection `mapstructure:"cloudflare"`
}

type MainSection struct {
	Listen                string `mapstructure:"listen"`                   // 监听地址
	DebugLevel            string `mapstructure:"debug-level"`              // 日志等级
	ChatGPTAPIPrefix      string `mapstructure:"ChatGPT_API_PREFIX"`       // GhatGPT网址前缀
	Endpoint              string `mapstructure:"endpoint"`                 // 后端服务器地址
	EnableVerifySharePage bool   `mapstructure:"enable-verify-share-page"` // 是否启用分享页验证
	EnableDayAPIPrefix    bool   `mapstructure:"enable-day-api-prefix"`    // 启用日抛域名支持
}
type CloudflareSection struct {
	Email  string `mapstructure:"email"`
	APIKey string `mapstructure:"api_key"`
	ZoneID string `mapstructure:"zone_id"`
	Notes  string `mapstructure:"notes"`
}

var Conf = new(VariableConfig)

// ReadConfig 读取配置文件
func ReadConfig() (*VariableConfig, error) {
	cmdViper := viper.New()  //读取命令行
	fileViper := viper.New() // 读取配置文件
	// 绑定命令行参数
	pflag.StringP("config", "c", "", "config file path")
	pflag.StringP("server", "s", ":8080", "server address")
	pflag.String("CHATGPT_API_PREFIX", "https://ai.fakeopen.com", "CHATGPT_API_PREFIX")
	pflag.String("debug-level", "info", "debug level")
	pflag.StringP("endpoint", "e", "http://127.0.0.1:8899", "endpoint")
	pflag.Bool("enable-verify-share-page", true, "enable verify share page")
	pflag.Bool("enable-day-api-prefix", true, "enable day api prefix")
	pflag.Parse()

	// 初始化Viperr
	err := cmdViper.BindPFlags(pflag.CommandLine)
	if err != nil {
		return nil, fmt.Errorf("iper.BindPFlags failed")
	}

	configFilePath := cmdViper.GetString("config")
	if configFilePath != "" {
		// 设置配置文件的名称和路径
		fileViper.SetConfigFile(configFilePath)
		// 设置配置文件的后缀（如果需要的话）
		fileViper.SetConfigType("yaml")

		// 读取配置文件
		if err := fileViper.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read config file: %v", err)
		}
		err := cmdViper.MergeConfigMap(fileViper.AllSettings())
		// 将配置文件的值合并到cmdViper实例中，但命令行参数的值仍具有较高优先级
		if err != nil {
			return nil, fmt.Errorf("failed to merge config file: %v", err)
		}
	}
	var ret = new(VariableConfig)
	// 使用cmdViper读取配置值
	if err := cmdViper.Unmarshal(ret); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config to struct: %v", err)
	}

	return ret, nil
}
