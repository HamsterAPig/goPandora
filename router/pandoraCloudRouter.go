package router

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	controller2 "goPandora/controller"
	logger "goPandora/internal/log"
	"goPandora/model"
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
	router.LoadHTMLGlob("resource/templates/*")

	// 加载静态文件
	router.Static("/css", "resource/static/_next/static/css")
	router.Static("/chunks", "resource/static/_next/static/chunks")
	router.Static("/"+model.Param.BuildId, "resource/static/_next/static/"+model.Param.BuildId)
	router.Static("/_next/static", "resource/static/_next/static")
	router.Static("/fonts", "resource/static/fonts")
	router.Static("/ulp", "resource/static/ulp")
	router.Static("/sweetalert2", "resource/static/sweetalert2")
	router.StaticFile("/service-worker.js", "resource/static/service-worker.js")
	router.StaticFile("/apple-touch-icon.png", "resource/static/apple-touch-icon.png")
	router.StaticFile("/favicon-16x16.png", "resource/static/favicon-16x16.png")
	router.StaticFile("/favicon-32x32.png", "resource/static/favicon-32x32.png")
	router.StaticFile("/manifest.json", "esource/static/manifest.json")
	router.StaticFile("/favicon.ico", "resource/static/favicon-16x16.png")

	// 404
	router.NoRoute(controller2.NotFoundHandler)
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
		api.GET("/auth/session", controller2.SessionAPIHandler)
		api.GET("/accounts/check/v4-2023-04-27", controller2.CheckAPIHandler)
	}

	router.GET("/", controller2.ChatHandler)
	chat := router.Group("/c")
	{
		chat.GET("", controller2.ChatHandler)
		chat.GET("/:chatID", controller2.ChatHandler)
	}

	_next := router.Group("/_next/data")
	{
		_next.GET(fmt.Sprintf("/%s/index.json", model.Param.BuildId), controller2.UserInfoHandler)
		_next.GET(fmt.Sprintf("/%s/c/:conversationID", model.Param.BuildId), controller2.UserInfoHandler)
		_next.GET(fmt.Sprintf("/%s/share/:shareID", model.Param.BuildId), controller2.ShareInfoHandler)
		_next.GET(fmt.Sprintf("/%s/share/:shareID/continue.json", model.Param.BuildId), controller2.ShareContinueHandler)
		_next.GET(fmt.Sprintf("/%s/auth/login.json", model.Param.BuildId), func(c *gin.Context) {
			c.JSON(http.StatusNotFound, gin.H{
				"status": "not found",
			})
		})
	}

	share := router.Group("/share")
	{
		share.GET("/:shareID", controller2.ShareDetailHandler)
		share.GET("/:shareID/continue", controller2.ShareContinueRedirect)
	}

	router.GET("/login", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/auth/login")
	})
	auth := router.Group("/auth")
	{
		auth.POST("/login_token", controller2.PostTokenHandler)
		auth.Any("/login", controller2.AuthLoginHandler)
		auth.GET("/logout", controller2.AuthLogoutHandler)
		auth.GET("/login_auto/:uuid", controller2.AutoLoginHandler)
		auth.GET("/login_share_token/:share_token", controller2.AuthLoginShareToken)
	}
	return router
}
