package config

import (
	"fmt"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type VariableConfig struct {
	MainConfig MainSection `mapstructure:"main"`
	WebConfig  WebSection  `mapstructure:"web"`
}

type MainSection struct {
	Listen           string   `mapstructure:"listen"`             // 监听地址
	AdminListen      string   `mapstructure:"admin_listen"`       // 管理员监听地址
	ProxyGroup       []string `mapstructure:"proxys"`             // 代理组
	DatabasePath     string   `mapstructure:"database"`           // 数据库路径
	DebugLevel       string   `mapstructure:"debug-level"`        // 日志等级
	ChatGPTAPIPrefix string   `mapstructure:"ChatGPT_API_PREFIX"` // GhatGPT网址前缀

	// 仅限命令行下使用的参数
	UserAddByFilePath string `mapstructure:"UserAddByFilePath"` // 通过文件批量添加用户
	ShareTokenAddByID bool   //通过user table id添加用户
	UserList          bool   `mapstructure:"UserList"` // 显示所有用户信息
	UserAdd           bool   `mapstructure:"UserAdd"`  // 添加用户
}

type WebSection struct {
	UserListPath      string `mapstructure:"WebUserListPath"`
	EnableSharePage   bool   `mapstructure:"EnableSharePageVerify"`
	WebsiteDomainName string `mapstructure:"WebSiteDomainName"`
}

var Conf = new(VariableConfig)

// ReadConfig 读取配置文件
func ReadConfig() (*VariableConfig, error) {
	cmdViper := viper.New()  //读取命令行
	fileViper := viper.New() // 读取配置文件
	// 绑定命令行参数
	pflag.StringP("config", "c", "", "config file path")
	pflag.StringP("server", "s", ":8080", "server address")
	pflag.StringSliceP("proxys", "p", nil, "proxy address")
	pflag.StringP("database", "b", "./data.db", "database file path")
	pflag.String("CHATGPT_API_PREFIX", "https://ai.fakeopen.com", "CHATGPT_API_PREFIX")
	pflag.String("user-add-file", "", "add user file path")
	pflag.String("web-user-list", "", "user list file path")
	pflag.String("debug-level", "info", "debug level")
	pflag.Bool("share-token-add-id", false, "share token add id")
	pflag.Bool("user-add", false, "add user")
	pflag.Bool("user-list", false, "list user")
	pflag.Bool("enable_share_page_verify", true, "enable share page verify")
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
	ret.MainConfig.UserList = cmdViper.GetBool("user-list")
	ret.MainConfig.UserAddByFilePath = cmdViper.GetString("user-add-file")
	ret.MainConfig.UserAdd = cmdViper.GetBool("user-add")
	ret.MainConfig.ShareTokenAddByID = cmdViper.GetBool("share-token-add-id")
	// 如果未指定配置文件且未提供其他命令行参数，则显示帮助信息
	if configFilePath == "" && !ret.MainConfig.UserAdd && !ret.MainConfig.UserList && ret.MainConfig.UserAddByFilePath == "" {
		pflag.Usage()
		return nil, fmt.Errorf("no enough arguments")
	}

	return ret, nil
}
