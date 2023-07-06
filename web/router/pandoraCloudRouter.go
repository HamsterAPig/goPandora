package router

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	logger "goPandora/internal/log"
	"goPandora/web/controller"
	"goPandora/web/model"
	"html/template"
	"net/http"
	"strings"
)

func PandoraCloudRouter() http.Handler {
	router := gin.Default()
	router.Delims("{[", "]}")
	// 注册自定义模板函数
	router.SetFuncMap(template.FuncMap{
		"safe": func(s string) template.HTML {
			return template.HTML(s)
		},
		"lower": func(value interface{}) interface{} {
			switch v := value.(type) {
			case string:
				return strings.ToLower(v)
			default:
				return v
			}
		},
		"tojson": func(v interface{}) template.JS {
			bytes, err := json.Marshal(v)
			if err != nil {
				logger.Error("json.Marshal failed", zap.Error(err))
				return ""
			}
			return template.JS(bytes)
		},
	})

	// 加载模板
	router.LoadHTMLGlob("web/gin/templates/*")

	// 加载静态文件
	router.Static("/_next/static", "web/gin/static/_next/static")
	router.Static("/fonts", "web/gin/static/fonts")
	router.Static("/ulp", "web/gin/static/ulp")
	router.Static("/sweetalert2", "web/gin/static/sweetalert2")
	router.StaticFile("/service-worker.js", "web/gin/static/service-worker.js")
	router.StaticFile("/apple-touch-icon.png", "web/gin/static/apple-touch-icon.png")
	router.StaticFile("/favicon-16x16.png", "web/gin/static/favicon-16x16.png")
	router.StaticFile("/favicon-32x32.png", "web/gin/static/favicon-32x32.png")
	router.StaticFile("/manifest.json", "web/gin/static/manifest.json")
	router.StaticFile("/favicon.ico", "web/gin/static/favicon-16x16.png")

	// 404
	router.NoRoute(controller.NotFoundHandler)
	router.GET("/404.html", func(c *gin.Context) {
		c.HTML(http.StatusOK, "404.html", gin.H{
			"api_prefix":     model.Param.ApiPrefix,
			"build_id":       model.Param.BuildId,
			"pandora_sentry": model.Param.PandoraSentry,
			"props":          "",
		})
	})

	api := router.Group("/api")
	{
		api.GET("/auth/session", controller.SessionAPIHandler)
		api.GET("/accounts/check/v4-2023-04-27", controller.CheckAPIHandler)
	}

	router.GET("/", controller.ChatHandler)
	chat := router.Group("/c")
	{
		chat.GET("", controller.ChatHandler)
		chat.GET("/:chatID", controller.ChatHandler)
	}

	_next := router.Group("/_next/data")
	{
		_next.GET(fmt.Sprintf("/%s/index.json", model.Param.BuildId), controller.UserInfoHandler)
		_next.GET(fmt.Sprintf("/%s/c/:conversationID", model.Param.BuildId), controller.UserInfoHandler)
		_next.GET(fmt.Sprintf("/%s/share/:shareID", model.Param.BuildId), controller.ShareInfoHandler)
		_next.GET(fmt.Sprintf("/%s/share/:shareID/continue.json", model.Param.BuildId), controller.ShareContinueHandler)
	}

	share := router.Group("/share")
	{
		share.GET("/:shareID", controller.ShareDetailHandler)
		share.GET("/:shareID/continue", controller.ShareContinueRedirect)
	}

	router.GET("/login", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/auth/login")
	})
	auth := router.Group("/auth")
	{
		auth.POST("/login_token", controller.PostTokenHandler)
		auth.Any("/login", controller.AuthLoginHandler)
		auth.GET("/logout", controller.AuthLogoutHandler)
		auth.GET("/login_auto/:uuid", controller.AutoLoginHandler)
		auth.GET("/login_share_token/:share_token", controller.AuthLoginShareToken)
	}
	return router
}
