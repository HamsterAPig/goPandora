package web

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"goPandora/config"
	logger "goPandora/internal/log"
	"goPandora/web/utils"
	"golang.org/x/sync/errgroup"
	"net/http"
	"time"
)

var (
	g errgroup.Group
)

func ServerStart() {
	// 设置gin日志等级
	if config.Conf.MainConfig.DebugLevel == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	if config.Conf.MainConfig.EnableDayAPIPrefix {
		serializedDate := time.Now().Format("20060102")
		config.Conf.MainConfig.ChatGPTAPIPrefix = fmt.Sprintf("https://ai-%s.fakeopen.com", serializedDate)
	}

	utils.Param = utils.PandoraParam{
		ApiPrefix:     config.Conf.MainConfig.ChatGPTAPIPrefix,
		PandoraSentry: false,
		BuildId:       "cx416mT2Lb0ZTj5FxFg1l",
	}
	// 设置是否启用分享页查看验证
	utils.Param.EnableSharePageVerify = config.Conf.MainConfig.EnableVerifySharePage
	// 启动服务
	pandoraCloud := &http.Server{
		Addr:         config.Conf.MainConfig.Listen,
		Handler:      utils.PandoraCloudRouter(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	g.Go(func() error {
		return pandoraCloud.ListenAndServe()
	})

	if err := g.Wait(); err != nil {
		logger.Fatal("ListenAndServe: ", zap.Error(err))
	}
}
