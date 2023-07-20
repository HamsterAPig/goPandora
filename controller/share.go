package controller

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"goPandora/model"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func ShareContinueRedirect(c *gin.Context) {
	c.Redirect(http.StatusPermanentRedirect, url.PathEscape(c.Param("/share/"+"shareID")))
}

// ShareDetailHandler 显示分享详情页
func ShareDetailHandler(c *gin.Context) {
	if model.Param.EnableSharePageVerify {
		_, _, _, _, err := getUserInfo(c)
		if err != nil {
			c.Redirect(http.StatusFound, "/auth/login?next=%2Fshare%2F"+c.Param("shareID"))
			c.Abort()
			return
		}
	}
	props := shareDetailJson(c)
	c.HTML(http.StatusOK, "share.html", gin.H{
		"props":      props,
		"api_prefix": model.Param.ApiPrefix,
	})
}

// shareDetailJson 返回反序列化之后的分享详情
func shareDetailJson(c *gin.Context) gin.H {
	shareID := c.Param("shareID")
	if strings.HasSuffix(shareID, ".json") {
		shareID = strings.TrimSuffix(shareID, ".json")
	}
	shareDetail, err := fetchShareDetail(shareID)
	if err != nil {
		error404(c)
	}
	_, exists := shareDetail["continue_conversation_url"]
	if exists {
		shareDetail["continue_conversation_url"] = strings.Replace(shareDetail["continue_conversation_url"].(string), "https://chat.openai.com", "", 1)
	}
	props := gin.H{
		"props": gin.H{
			"pageProps": gin.H{
				"sharedConversationId": c.Param("shareID"),
				"serverResponse": gin.H{
					"type": "data",
					"data": shareDetail,
				},
				"continueMode":   false,
				"moderationMode": false,
				"chatPageProps":  gin.H{},
			},
			"__N_SSP": true,
		},
		"page": "/share/[[...shareParams]]",
		"query": gin.H{
			"shareParams": []string{c.Param("shareID")},
		},
		"buildId":      model.Param.BuildId,
		"isFallback":   false,
		"gssp":         true,
		"scriptLoader": []string{},
	}
	return props
}

// fetchShareDetail 从源服务器处抓取share info
func fetchShareDetail(shareID string) (retJson map[string]interface{}, err error) {
	url1 := model.Param.ApiPrefix + "/api/share/" + shareID
	resp, err := http.Get(url1)
	if err != nil {
		return
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body error: %s", err)
	}
	jsonStr := string(body)
	var data map[string]interface{}
	err = json.Unmarshal([]byte(jsonStr), &data)
	if err != nil {
		return nil, fmt.Errorf("json unmarshal error: %s", err)
	}
	return data, nil
}

// ShareInfoHandler 以JSON的格式返回分享页详情
func ShareInfoHandler(c *gin.Context) {
	props := shareDetailJson(c)
	c.JSON(http.StatusOK, props)
}
func ShareContinueHandler(c *gin.Context) {
	// 检查是否登陆，未登录则返回登陆Url
	userID, email, _, _, err := getUserInfo(c)
	if err != nil {
		nextURL := fmt.Sprintf("/share/%s/continue", url.PathEscape(c.Param("shareID")))
		loginURL := fmt.Sprintf("/auth/login?next=%s", url.PathEscape(nextURL))
		c.JSON(http.StatusOK, gin.H{
			"pageProps": gin.H{
				"__N_REDIRECT":        loginURL,
				"__N_REDIRECT_STATUS": http.StatusTemporaryRedirect,
			},
			"__N_SSP": true,
		})
		c.Abort()
		return
	}
	shareID := c.Param("shareID")
	shareDetail, err := fetchShareDetail(shareID)
	if err != nil {
		error404(c)
	}
	_, exists := shareDetail["continue_conversation_url"]
	if exists {
		shareDetail["continue_conversation_url"] = strings.Replace(shareDetail["continue_conversation_url"].(string), "https://chat.openai.com", "", 1)
	}
	props := gin.H{
		"pageProps": gin.H{
			"user": gin.H{
				"id":      userID,
				"name":    email,
				"email":   email,
				"image":   nil,
				"picture": nil,
				"groups":  []string{},
			},
			"serviceStatus": gin.H{},
			"userCountry":   "US",
			"geoOk":         true,
			"serviceAnnouncement": gin.H{
				"paid":   gin.H{},
				"public": gin.H{},
			},
			"isUserInCanPayGroup":  true,
			"sharedConversationId": shareID,
			"serverResponse": gin.H{
				"type": "data",
				"data": shareDetail,
			},
			"continueMode":   true,
			"moderationMode": false,
			"chatPageProps": gin.H{
				"user": gin.H{
					"id":      userID,
					"name":    email,
					"email":   email,
					"image":   nil,
					"picture": nil,
					"groups":  []string{},
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
		},
		"__N_SSP": true,
	}
	c.JSON(http.StatusOK, props)
}
