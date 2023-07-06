package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"goPandora/internal/pandora"
	"goPandora/web/model"
	"net/http"
	"strings"
	"time"
)

// SessionAPIHandler 获取用户信息接口
func SessionAPIHandler(c *gin.Context) {
	// 解析、验证token并且返回值
	userID, email, accessToken, payload, err := getUserInfo(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		c.Abort()
		return
	}

	// 序列化payload的过期时间
	exp, ok := payload["exp"].(float64)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid token because exp is not float64",
		})
		c.Abort()
		return
	}
	expTimestamp := time.Unix(int64(exp), 0).Format("2006-01-02T15:04:05")

	// 构造返回json
	ret := &gin.H{
		"user": gin.H{
			"id":      userID,
			"name":    email,
			"email":   email,
			"image":   nil,
			"picture": nil,
			"groups":  []interface{}{},
		},
		"expires":      expTimestamp,
		"accessToken":  accessToken,
		"authProvider": "auth0",
	}
	c.JSON(http.StatusOK, ret)
}

// PostTokenHandler 使用token登陆
func PostTokenHandler(c *gin.Context) {
	// 从post数据中获取next url
	next := c.PostForm("next")
	if "" == next { // 如果获取到的next url为空的话则设置为/
		next = "/"
	}
	// 从post数据中获取access-token
	accessToken := c.PostForm("access_token")
	if "" != accessToken {
		// 检查access-token
		var payload jwt.MapClaims
		var err error
		if strings.HasPrefix(accessToken, "fk-") {
			var info model.ShareTokenResponse
			info, err = fetchShareTokenInfo(accessToken)
			payload = jwt.MapClaims{"exp": float64(info.ExpireAt)}
		} else {
			payload, err = pandora.CheckAccessToken(accessToken)
		}
		if nil != err {
			data := gin.H{"code": 1, "msg": err.Error()}
			c.JSON(http.StatusInternalServerError, data)
			c.Abort()
			return
		}

		// 检查token的过期时间
		exp, _ := payload["exp"].(float64)
		expires := time.Unix(int64(exp), 0)

		data := gin.H{"code": 0, "url": next} // 返回状态

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
		c.JSON(http.StatusOK, data)
		c.Redirect(http.StatusFound, next)
	} else { // 错误的access-token处理
		data := gin.H{"code": 1, "msg": "access token is null"}
		c.JSON(http.StatusInternalServerError, data)
	}
}
