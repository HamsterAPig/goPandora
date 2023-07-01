package controller

import (
	"github.com/gin-gonic/gin"
	"goPandora/web/model"
	"net/http"
)

// ChatHandler 主入口函数
func ChatHandler(c *gin.Context) {
	conversationID := c.Param("chatID")
	// 解析、验证token
	userID, email, _, _, err := getUserInfo(c)
	if err != nil { // 如果验证的token出现错误则跳转到/login
		c.Redirect(http.StatusFound, "/login")
		c.Abort()
		return
	}

	// 构造返回json
	props := gin.H{
		"props": gin.H{
			"pageProps": gin.H{
				"user": gin.H{
					"id":      userID,
					"name":    email,
					"email":   email,
					"image":   nil,
					"picture": nil,
					"groups":  []interface{}{},
				},
				"serviceStatus": gin.H{},
				"userCountry":   "US",
				"geoOk":         true,
				"serviceAnnouncement": gin.H{
					"paid":   gin.H{},
					"public": gin.H{},
				},
				"isUserInCanPayGroup": true,
			},
			"__N_SSP": true,
		},
		"page":         "/",
		"query":        gin.H{},
		"buildId":      model.Param.BuildId,
		"isFallback":   false,
		"gssp":         true,
		"scriptLoader": []interface{}{},
	}

	// 根据会话ID选择模板
	templateHtml := "chat.html"
	if "" != conversationID {
		templateHtml = "detail.html"
		props["page"] = "/c/[chatId]"
		props["query"] = gin.H{"chatId": conversationID}
	}

	// 返回渲染好的模板
	c.HTML(http.StatusOK, templateHtml, gin.H{
		"pandora_sentry": model.Param.PandoraSentry,
		"api_prefix":     model.Param.ApiPrefix,
		"props":          props,
	})
}
