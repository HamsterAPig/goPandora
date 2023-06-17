package utils

import (
	"encoding/json"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"goPandora/internal/db"
	logger "goPandora/internal/log"
	"html/template"
	"net/http"
	"strings"
)

func AdminRouter() http.Handler {
	router := gin.Default()
	router.Delims("{[", "]}")
	router.Use(cors.Default())

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

	router.GET("/list-user-all", func(c *gin.Context) {
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
	v1 := router.Group("/api/v1")
	{
		v1.GET("/user-list", func(c *gin.Context) {
			ret, err := db.GetAllUserInfo()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
			}
			c.JSON(http.StatusOK, ret)
		})
		v1.GET("/share-token-list", func(c *gin.Context) {
			ret, err := db.GetAllAutoLoginInfo()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
			}
			c.JSON(http.StatusOK, ret)
		})
	}
	return router
}
