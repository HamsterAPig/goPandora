package utils

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"goPandora/config"
	"goPandora/internal/db"
	logger "goPandora/internal/log"
	"html/template"
	"net/http"
	"strings"
)

func AdminRouter() http.Handler {
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
	router.LoadHTMLGlob("web/gin/admin/templates/*")

	router.GET(config.Conf.WebConfig.UserListPath, func(c *gin.Context) {
		ret := db.ListAllUser()
		// 构建二维切片，每个元素是字符串分割后的结果
		var data [][]string
		for _, str := range ret {
			parts := strings.Split(str, ",")
			data = append(data, parts)
		}
		c.HTML(http.StatusOK, "list_user.html", gin.H{
			"userList": data,
		})
	})

	apiV1Group := router.Group("/api/v1")
	{
		// 显示所有的ShareToken
		apiV1Group.GET("/getAllShareToken", func(c *gin.Context) {
			shareTokens, err := db.GetAllShareToken()
			if err != nil {
				c.JSON(http.StatusOK, gin.H{
					"error": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, shareTokens)
		})

		// 通过userID获取accessToken
		apiV1Group.GET("/getAccessToken", func(c *gin.Context) {
			userID := c.Query("userID")
			accessToken, err := db.GetAccessTokenByUserID(userID)
			if err != nil {
				c.JSON(http.StatusOK, gin.H{
					"error": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, accessToken)
		})
	}
	return router
}
