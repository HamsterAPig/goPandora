package web

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"go.uber.org/zap"
	logger "goPandora/internal/log"
	"goPandora/internal/pandora"
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
		"tojson": func(v interface{}) string {
			jsonData, err := json.Marshal(v)
			if err != nil {
				logger.Error("json.Marshal failed", zap.Error(err))
				return ""
			}
			return string(jsonData)
		},
	})
	// 加载模板
	router.LoadHTMLGlob("web/gin/templates/*")

	// 加载静态文件
	router.Static("/_next/static", "web/gin/static/_next/static")
	router.Static("/fonts", "web/gin/static/fonts")
	router.Static("/ulp", "web/gin/static/ulp")
	router.StaticFile("/service-worker.js", "web/gin/static/service-worker.js")
	router.StaticFile("/apple-touch-icon.png", "web/gin/static/apple-touch-icon.png")
	router.StaticFile("/favicon-16x16.png", "web/gin/static/favicon-16x16.png")
	router.StaticFile("/favicon-32x32.png", "web/gin/static/favicon-32x32.png")
	router.StaticFile("/manifest.json", "web/gin/static/manifest.json")
	router.StaticFile("/favicon.ico", "web/gin/static/favicon.ico")

	// 配置路由
	router.GET("/api/auth/session", sessionAPIHandler)
	router.GET("/api/accounts/check/v4-2023-04-27", checkAPIHandler)
	router.GET(fmt.Sprintf("/_next/data/%s/index.json", param.BuildId), userInfoHandler)
	router.GET(fmt.Sprintf("/_next/data/%s/c/:conversationID", param.BuildId), userInfoHandler)

	router.GET("/", func(c *gin.Context) {
		chatHandler(c, param, "")
	})
	router.GET("/c", func(c *gin.Context) {
		chatHandler(c, param, "")
	})
	router.GET("/c/:chatID", func(c *gin.Context) {
		chatHandler(c, param, c.Param("chatID"))
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
	router.GET("/auth/logout", func(context *gin.Context) {
		context.SetCookie("access-token", "", -1, "/", "", false, true)
		context.Redirect(http.StatusMovedPermanently, "/auth/login")
	})

	// 启动服务
	err := router.Run(address)
	if err != nil {
		return
	}
}

func userInfoHandler(c *gin.Context) {
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

// checkAPIHandler 返回一组json用来给当前账号授予一些网页版的视觉上面的特性
func checkAPIHandler(c *gin.Context) {
	// 下面这组数据是从Pandora中直接拿出来的
	data := gin.H{
		"accounts": gin.H{
			"default": gin.H{
				"account": gin.H{
					"account_user_role": "account-owner",
					"account_user_id":   "d0322341-7ace-4484-b3f7-89b03e82b927",
					"processor": gin.H{
						"a001": gin.H{
							"has_customer_object": true,
						},
						"b001": gin.H{
							"has_transaction_history": true,
						},
					},
					"account_id": "a323bd05-db25-4e8f-9173-2f0c228cc8fa",
					"is_most_recent_expired_subscription_gratis": true,
					"has_previously_paid_subscription":           true,
				},
				"features": []string{
					"model_switcher",
					"model_preview",
					"system_message",
					"data_controls_enabled",
					"data_export_enabled",
					"show_existing_user_age_confirmation_modal",
					"bucketed_history",
					"priority_driven_models_list",
					"message_style_202305",
					"layout_may_2023",
					"plugins_available",
					"beta_features",
					"infinite_scroll_history",
					"browsing_available",
					"browsing_inner_monologue",
					"browsing_bing_branding",
					"shareable_links",
					"plugin_display_params",
					"tools3_dev",
					"tools2",
					"debug",
				},
				"entitlement": gin.H{
					"subscription_id":         "d0dcb1fc-56aa-4cd9-90ef-37f1e03576d3",
					"has_active_subscription": true,
					"subscription_plan":       "chatgptplusplan",
					"expires_at":              "2089-08-08T23:59:59+00:00",
				},
				"last_active_subscription": gin.H{
					"subscription_id":          "d0dcb1fc-56aa-4cd9-90ef-37f1e03576d3",
					"purchase_origin_platform": "chatgpt_mobile_ios",
					"will_renew":               true,
				},
			},
		},
		"temp_ap_available_at": "2023-05-20T17:30:00+00:00",
	}
	c.JSON(http.StatusOK, data)
}

// sessionAPIHandler 获取用户信息接口
func sessionAPIHandler(c *gin.Context) {
	// 解析、验证token并且返回值
	userID, email, accessToken, payload, err := getUserInfo(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
	}

	// 序列化payload的过期时间
	exp, ok := payload["exp"].(float64)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid token because exp is not float64",
		})
	}
	expTimestamp := time.Unix(int64(exp), 0).Format("2006-01-02 15:04:05")

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

// chatHandler 主入口函数
func chatHandler(ctx *gin.Context, param *PandoraParam, conversationID string) {
	// 解析、验证token
	userID, email, _, _, err := getUserInfo(ctx)
	if err != nil { // 如果验证的token出现错误则跳转到/login
		ctx.Redirect(http.StatusFound, "/login")
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
		"query":        "{}",
		"buildId":      param.BuildId,
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
	ctx.HTML(http.StatusOK, templateHtml, gin.H{
		"pandora_sentry": param.PandoraSentry,
		"api_prefix":     param.ApiPrefix,
		"props":          props,
	})
}

// postTokenHandler 使用token登陆
func postTokenHandler(c *gin.Context) {
	// 从post数据中获取next url
	next := c.PostForm("next")
	if "" == next { // 如果获取到的next url为空的话则设置为/
		next = "/"
	}
	// 从post数据中获取access-token
	accessToken := c.PostForm("access_token")
	if "" != accessToken {
		// 检查access-token
		payload, err := pandora.CheckAccessToken(accessToken)
		if nil != err {
			data := gin.H{"code": 1, "msg": err.Error()}
			c.JSON(http.StatusInternalServerError, data)
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
	} else { // 错误的access-token处理
		data := gin.H{"code": 1, "msg": "access token is null"}
		c.JSON(http.StatusInternalServerError, data)
	}
}

// getUserInfo 从token获取用户信息
func getUserInfo(c *gin.Context) (string, string, string, jwt.MapClaims, error) {
	accessToken, err := c.Cookie("access-token")
	if err != nil {
		return "", "", "", nil, err
	}
	return pandora.CheckUserInfo(accessToken)
}
