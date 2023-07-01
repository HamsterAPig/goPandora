package controller

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"goPandora/config"
	logger "goPandora/internal/log"
	"goPandora/internal/pandora"
	"goPandora/web/model"
	"io"
	"net/http"
	"strings"
	"time"
)

// AuthLoginHandler 官方账号密码登陆
func AuthLoginHandler(c *gin.Context) {
	switch c.Request.Method {
	case http.MethodPost:
		userName := c.PostForm("username")
		password := c.PostForm("password")
		//mfaCode := c.PostForm("mfa_code")
		nextUrl := c.PostForm("next")

		if userName != "" && password != "" {
			accessToken, _, err := pandora.Auth0(userName, password, "", "")
			if err != nil {
				c.HTML(http.StatusOK, "login.html", gin.H{
					"username": userName,
					"error":    err.Error(),
				})
			}
			payload, err := pandora.CheckAccessToken(accessToken)
			if err != nil {
				c.HTML(http.StatusOK, "login.html", gin.H{
					"username": userName,
					"error":    err.Error(),
				})
			}
			// 检查token的过期时间
			exp, _ := payload["exp"].(float64)
			expires := time.Unix(int64(exp), 0)

			// 设置cookie
			cookie := &http.Cookie{
				Name:     "access-token",
				Value:    accessToken,
				Expires:  expires,
				Path:     "/",
				Domain:   "",
				Secure:   false,
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
			}
			http.SetCookie(c.Writer, cookie)
			if nextUrl == "" {
				nextUrl = "/"
			}
			c.Redirect(http.StatusFound, nextUrl)
		}
	default:
		next := c.Query("next")
		c.HTML(http.StatusOK, "login.html", gin.H{
			"api_prefix": model.Param.ApiPrefix,
			"next":       next,
		})
	}
}

func AuthLogoutHandler(c *gin.Context) {
	c.SetCookie("access-token", "", -1, "/", "", false, true)
	c.Redirect(http.StatusFound, "/auth/login")
}

// AutoLoginHandler 在访问自动登陆页面时自动设置cookie
func AutoLoginHandler(c *gin.Context) {
	uuid := c.Param("uuid")
	// 发送GET请求
	resp, err := http.Get(config.Conf.MainConfig.Endpoint + uuid)
	if err != nil {
		c.String(http.StatusInternalServerError, "发送Get请求失败")
		c.Abort()
		return
	}
	defer resp.Body.Close()

	// 读取响应内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("读取响应内容失败", zap.Error(err))
		c.String(http.StatusInternalServerError, "\n读取响应内容失败")
		c.Abort()
		return
	}

	// 解析JSON数据为map
	var responseData map[string]interface{}
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		logger.Error("解析JSON数据失败", zap.Error(err))
		c.String(http.StatusInternalServerError, "\n解析JSON数据失败")
		c.Abort()
		return
	}

	// 提取Token值
	token, ok := responseData["data"].(map[string]interface{})["Token"].(string)
	if !ok {
		logger.Error("提取Token值失败", zap.Error(err))
		c.String(http.StatusInternalServerError, "\n"+"提取Token值失败")
		c.Abort()
		return
	}
	payload, err := pandora.CheckAccessToken(token)
	if err != nil {
		logger.Error("pandora.GetTokenByRefreshToken failed", zap.Error(err))
		c.String(http.StatusInternalServerError, "\n"+"pandora.GetTokenByRefreshToken failed")
		c.Abort()
		return
	}
	exp, _ := payload["exp"].(float64)
	expires := time.Unix(int64(exp), 0)
	// 设置cookie
	cookie := &http.Cookie{
		Name:     "access-token",
		Value:    token,
		Expires:  expires,
		Path:     "/",
		Domain:   "",
		Secure:   false,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(c.Writer, cookie)
	c.Redirect(http.StatusFound, "/")
	c.String(http.StatusOK, "\n若网页并没有跳转，请手动刷新本页...")
}

func AuthLoginShareToken(c *gin.Context) {
	shareToken := c.Param("share_token")
	if shareToken != "" && strings.HasPrefix(shareToken, "fk-") {
		info, err := fetchShareTokenInfo(shareToken)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": err.Error(),
			})
			c.Abort()
			return
		}
		// 设置cookie
		cookie := &http.Cookie{
			Name:     "access-token",
			Value:    shareToken,
			Expires:  time.Unix(info.ExpireAt, 0),
			Path:     "/",
			Domain:   "",
			Secure:   false,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		}
		http.SetCookie(c.Writer, cookie)
		c.Redirect(http.StatusFound, "/")
		c.String(http.StatusOK, "\n若网页并没有跳转，请手动刷新本页...")
	}
}

func fetchShareTokenInfo(token string) (model.ShareTokenResponse, error) {
	var shareInfo model.ShareTokenResponse
	shareInfo.ExpireAt = -1

	resp, err := http.Get(model.Param.ApiPrefix + "/token/info/" + token)
	if err != nil {
		return shareInfo, fmt.Errorf("获取token信息失败: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return shareInfo, fmt.Errorf("share token ot found or expired")
	}
	if resp.StatusCode != http.StatusOK {
		return shareInfo, fmt.Errorf("failed to fetch share token info")
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return shareInfo, fmt.Errorf("read body error: %s", err)
	}
	err = json.Unmarshal([]byte(body), &shareInfo)
	if err != nil {
		return shareInfo, fmt.Errorf("json unmarshal error: %s", err)
	}
	return shareInfo, nil
}
