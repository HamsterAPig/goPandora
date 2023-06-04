package web

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"go.uber.org/zap"
	logger "goPandora/internal/log"
	"html/template"
	"net/http"
	"strings"
	"time"
)

type PandoraParam struct {
	ApiPrefix     string
	PandoraSentry string
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

	router.LoadHTMLGlob("web/gin/templates/*")

	router.Static("/_next", "web/gin/static/_next")
	router.Static("/fonts", "web/gin/static/fonts")
	router.Static("/ulp", "web/gin/static/ulp")
	router.Static("/static", "web/gin/static")

	router.GET("/", func(c *gin.Context) {
		chatHandler(c, param, "")
	})
	router.GET("/login", func(context *gin.Context) {
		context.Redirect(http.StatusMovedPermanently, "/auth/login")
	})
	router.GET("/auth/login", func(context *gin.Context) {
		next := context.Query("next")
		context.HTML(http.StatusOK, "login.html", gin.H{
			"api_prefix": param.ApiPrefix,
			"next":       next,
		})
	})
	router.POST("/auth/login_token", postTokenHandler)
	err := router.Run(address)
	if err != nil {
		return
	}
}

func chatHandler(ctx *gin.Context, param *PandoraParam, conversationID string) {
	accessToken, _ := ctx.Cookie("access-token")
	userID, email, _, _, err := getUserInfo(accessToken)
	if err != nil {
		ctx.Redirect(http.StatusFound, "/login")
	}

	query := ctx.Request.URL.Query()
	if conversationID != "" {
		query.Set("chatId", conversationID)
	}

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
		"query":        query,
		"buildId":      param.BuildId,
		"isFallback":   false,
		"gssp":         true,
		"scriptLoader": []interface{}{},
	}
	//propsStr := fmt.Sprint(props)

	//templateHtml := "detail.html"
	//if conversationID == "" {
	//	templateHtml = "chat.html"
	//}
	ctx.JSON(http.StatusOK, props)
	//ctx.HTML(http.StatusOK, templateHtml, gin.H{
	//	"pandora_sentry": param.PandoraSentry,
	//	"api_prefix":     param.ApiPrefix,
	//	"props":          propsStr,
	//})
}

// postTokenHandler 使用token登陆
func postTokenHandler(c *gin.Context) {
	next := c.PostForm("next")
	if "" == next {
		next = "/"
	}
	accessToken := c.PostForm("access_token")
	if "" != accessToken {
		payload, err := CheckAccessToken(accessToken)
		if nil != err {
			data := gin.H{"code": 1, "msg": err.Error()}
			c.JSON(http.StatusInternalServerError, data)
		}
		exp, _ := payload["exp"].(float64)
		expires := time.Unix(int64(exp), 0)
		data := gin.H{"code": 0, "url": next}
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
	} else {
		data := gin.H{"code": 1, "msg": "access token is null"}
		c.JSON(http.StatusInternalServerError, data)
	}
}

// 从token获取用户信息
func getUserInfo(accessToken string) (string, string, string, jwt.MapClaims, error) {
	payload, err := CheckAccessToken(accessToken)
	if nil != err {
		logger.Error("CheckAccessToken failed", zap.Error(err))
	}
	// 使用类型断言访问声明中的属性
	var email, userID string
	if profile, ok := payload["https://api.openai.com/profile"].(map[string]interface{}); ok {
		if emailVal, ok := profile["email"].(string); !ok {
			return "", "", "", nil, fmt.Errorf("failed to get email")
		} else {
			email = emailVal
		}
	}

	if auth, ok := payload["https://api.openai.com/auth"].(map[string]interface{}); ok {
		if userIDVal, ok := auth["user_id"].(string); !ok {
			return "", "", "", nil, fmt.Errorf("failed to get user_id")
		} else {
			userID = userIDVal
		}
	}
	return userID, email, accessToken, payload, nil
}

// CheckAccessToken 检查token并且返回payload
func CheckAccessToken(accessToken string) (jwt.MapClaims, error) {
	publicKey := `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA27rOErDOPvPc3mOADYtQ
BeenQm5NS5VHVaoO/Zmgsf1M0Wa/2WgLm9jX65Ru/K8Az2f4MOdpBxxLL686ZS+K
7eJC/oOnrxCRzFYBqQbYo+JMeqNkrCn34yed4XkX4ttoHi7MwCEpVfb05Qf/ZAmN
I1XjecFYTyZQFrd9LjkX6lr05zY6aM/+MCBNeBWp35pLLKhiq9AieB1wbDPcGnqx
lXuU/bLgIyqUltqLkr9JHsf/2T4VrXXNyNeQyBq5wjYlRkpBQDDDNOcdGpx1buRr
Z2hFyYuXDRrMcR6BQGC0ur9hI5obRYlchDFhlb0ElsJ2bshDDGRk5k3doHqbhj2I
gQIDAQAB
-----END PUBLIC KEY-----`

	// 解析token
	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
		publicKey, err := jwt.ParseRSAPublicKeyFromPEM([]byte(publicKey))
		if nil != err {
			return nil, fmt.Errorf("failed to parse public key: %v", err)
		}
		return publicKey, nil
	})

	if nil != err {
		return nil, fmt.Errorf("failed to parse token: %v", err)
	}

	// 验证 JWT 的有效性
	if !token.Valid {
		return nil, fmt.Errorf("invalid JWT")
	}

	// 获取 payload
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("failed to get JWT claims")
	}
	if _, ok := claims["scope"]; !ok {
		return nil, fmt.Errorf("miss scope")
	}
	scope := claims["scope"]
	if !strings.Contains(scope.(string), "model.read") || !strings.Contains(scope.(string), "model.request") {
		return nil, fmt.Errorf("invalid scope")
	}
	_, ok1 := claims["https://api.openai.com/auth"]
	_, ok2 := claims["https://api.openai.com/profile"]
	if !ok1 || !ok2 {
		return nil, fmt.Errorf("belonging to an unregistered user")
	}

	return claims, nil
}
