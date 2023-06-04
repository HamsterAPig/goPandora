package web

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	logger "goPandora/internal/log"
	"html/template"
	"net/http"
	"strings"
)

type PandoraParam struct {
	ApiPrefix     string
	PandoraSentry bool
	BuildId       string
}

func ServerStart(address string, param *PandoraParam) {
	router := gin.Default()
	router.Delims("{[", "]}")
	// 注册自定义模板函数
	router.SetFuncMap(template.FuncMap{
		"safe": func(s string) template.HTML {
			return template.HTML(s)
		},
		"lower": strings.ToLower,
		"tojson": func(v interface{}) template.HTML {
			jsonData, err := json.Marshal(v)
			if err != nil {
				logger.Error("json.Marshal failed", zap.Error(err))
				return ""
			}
			return template.HTML(jsonData)
		},
	})

	router.LoadHTMLFiles("web/gin/templates/login.html")

	router.Static("/_next", "web/gin/static/_next")
	router.Static("/fonts", "web/gin/static/fonts")
	router.Static("/ulp", "web/gin/static/ulp")

	//router.GET("/404.html", func(c *gin.Context) {
	//	c.HTML(http.StatusOK, "404.html", gin.H{
	//		"pandora_sentry": "false",
	//		"api_prefix":     chatGPTAPI,
	//	})
	//})
	router.GET("/auth/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", gin.H{
			"pandora_sentry": param.PandoraSentry,
			"api_prefix":     param.ApiPrefix,
			"error":          "错误",
		})
	})
	err := router.Run(address)
	if err != nil {
		return
	}
}
