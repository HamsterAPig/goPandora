package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"goPandora/internal/pandora"
	"net/http"
	"strings"
)

// getUserInfo 从token获取用户信息
func getUserInfo(c *gin.Context) (string, string, string, jwt.MapClaims, error) {
	accessToken, err := c.Cookie("access-token")
	if err != nil {
		return "", "", "", nil, err
	}
	if strings.HasPrefix(accessToken, "fk-") {
		info, err := fetchShareTokenInfo(accessToken)
		if nil != err {
			return "", "", "", nil, err
		}
		return info.UserID, info.Email, accessToken, jwt.MapClaims{"exp": float64(info.ExpireAt)}, nil
	} else {
		return pandora.CheckUserInfo(accessToken)
	}
}

// UserInfoHandler 获取当前用户的信息
func UserInfoHandler(c *gin.Context) {
	userID, email, _, _, err := getUserInfo(c)
	if nil != err {
		data := gin.H{
			"pageProps": gin.H{
				"__N_REDIRECT":        "/auth/login?",
				"__N_REDIRECT_STATUS": 307,
			},
			"__N_SSP": true,
		}
		c.JSON(http.StatusBadRequest, data)
	}
	ret := gin.H{
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
	}
	c.JSON(http.StatusOK, ret)
}
