package main

import (
	"fmt"
	"goPandora/config"
	logger "goPandora/internal/log"
	"goPandora/web"
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
	web.ServerStart()
}
