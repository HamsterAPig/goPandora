package web

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"goPandora/config"
	logger "goPandora/internal/log"
	"goPandora/web/model"
	"goPandora/web/router"
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

	model.Param = model.PandoraParam{
		ApiPrefix:     config.Conf.MainConfig.ChatGPTAPIPrefix,
		PandoraSentry: false,
		BuildId:       "cx416mT2Lb0ZTj5FxFg1l",
	}
	// 设置是否启用分享页查看验证
	model.Param.EnableSharePageVerify = config.Conf.MainConfig.EnableVerifySharePage
	// 启动服务
	pandoraCloud := &http.Server{
		Addr:         config.Conf.MainConfig.Listen,
		Handler:      router.PandoraCloudRouter(),
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
