package web

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"go.uber.org/zap"
	"goPandora/config"
	"goPandora/internal/db"
	logger "goPandora/internal/log"
	"goPandora/internal/pandora"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type PandoraParam struct {
	ApiPrefix             string
	PandoraSentry         bool
	BuildId               string
	EnableSharePageVerify bool
}

var Param PandoraParam

func ServerStart(address string) {
	// 设置gin日志等级
	if config.Conf.MainConfig.DebugLevel == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

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
	router.LoadHTMLGlob("web/gin/templates/*")

	// 加载静态文件
	router.Static("/_next/static", "web/gin/static/_next/static")
	router.Static("/fonts", "web/gin/static/fonts")
	router.Static("/ulp", "web/gin/static/ulp")
	router.Static("/sweetalert2", "web/gin/static/sweetalert2")
	router.StaticFile("/service-worker.js", "web/gin/static/service-worker.js")
	router.StaticFile("/apple-touch-icon.png", "web/gin/static/apple-touch-icon.png")
	router.StaticFile("/favicon-16x16.png", "web/gin/static/favicon-16x16.png")
	router.StaticFile("/favicon-32x32.png", "web/gin/static/favicon-32x32.png")
	router.StaticFile("/manifest.json", "web/gin/static/manifest.json")
	router.StaticFile("/favicon.ico", "web/gin/static/favicon-16x16.png")

	// 配置路由
	router.GET("/api/auth/session", sessionAPIHandler)
	router.GET("/api/accounts/check/v4-2023-04-27", checkAPIHandler)
	router.GET(fmt.Sprintf("/_next/data/%s/index.json", Param.BuildId), userInfoHandler)
	router.GET(fmt.Sprintf("/_next/data/%s/c/:conversationID", Param.BuildId), userInfoHandler)
	router.GET(fmt.Sprintf("/_next/data/%s/share/:shareID", Param.BuildId), shareInfoHandler)
	router.GET(fmt.Sprintf("/_next/data/%s/share/:shareID/continue.json", Param.BuildId), shareContinueHandler)

	router.GET("/", chatHandler)
	router.GET("/c", chatHandler)
	router.GET("/c/:chatID", chatHandler)

	router.GET("/login", func(context *gin.Context) {
		context.Redirect(http.StatusFound, "/auth/login")
	})
	router.GET("/auth/login", func(context *gin.Context) {
		next := context.Query("next")
		context.HTML(http.StatusOK, "login.html", gin.H{
			"api_prefix": Param.ApiPrefix,
			"next":       next,
		})
	})
	router.POST("/auth/login_token", postTokenHandler)
	router.POST("/auth/login", postLoginHandler)
	router.GET("/auth/logout", func(context *gin.Context) {
		context.SetCookie("access-token", "", -1, "/", "", false, true)
		context.Redirect(http.StatusFound, "/auth/login")
	})

	// 自动设置cookie以到达访问url自动登陆的页面
	router.GET("/auth/login_auto/:uuid", autoLoginHandler)
	if config.Conf.WebConfig.UserListPath != "" {
		router.GET(config.Conf.WebConfig.UserListPath, func(c *gin.Context) {
			ret := db.ListAllUser()
			// 构建二维切片，每个元素是字符串分割后的结果
			data := [][]string{}
			for _, str := range ret {
				parts := strings.Split(str, ",")
				data = append(data, parts)
			}
			c.HTML(http.StatusOK, "list_user.html", gin.H{
				"userList": data,
			})
		})
	}

	// 404
	router.GET("/404.html", func(c *gin.Context) {
		c.HTML(http.StatusOK, "404.html", gin.H{
			"api_prefix":     Param.ApiPrefix,
			"build_id":       Param.BuildId,
			"pandora_sentry": Param.PandoraSentry,
			"props":          "",
		})
	})

	router.GET("/share/:shareID", shareDetailHandler)
	router.GET("/share/:shareID/continue", func(context *gin.Context) {
		context.Redirect(http.StatusPermanentRedirect, url.PathEscape(context.Param("/share/"+"shareID")))
	})

	// 启动服务
	err := router.Run(address)
	if err != nil {
		return
	}
}

func shareContinueHandler(c *gin.Context) {
	// 检查是否登陆，未登录则返回登陆Url
	userID, email, _, _, err := getUserInfo(c)
	if err != nil {
		nextURL := fmt.Sprintf("/share/%s/continue", url.PathEscape(c.Param("shareID")))
		loginURL := fmt.Sprintf("/auth/login?next=%s", url.PathEscape(nextURL))
		c.JSON(http.StatusForbidden, gin.H{
			"pageProps": gin.H{
				"__N_REDIRECT":        loginURL,
				"__N_REDIRECT_STATUS": true,
			},
			"__N_SSP": true,
		})
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

// shareInfoHandler 以JSON的格式返回分享页详情
func shareInfoHandler(c *gin.Context) {
	props := shareDetailJson(c)
	c.JSON(http.StatusOK, props)
}

// shareDetailHandler 显示分享详情页
func shareDetailHandler(c *gin.Context) {
	if Param.EnableSharePageVerify {
		_, _, _, _, err := getUserInfo(c)
		if err != nil {
			c.Redirect(http.StatusFound, "/auth/login?next=%2Fshare%2F"+c.Param("shareID"))
			return
		}
	}
	props := shareDetailJson(c)
	c.HTML(http.StatusOK, "share.html", gin.H{
		"props":          props,
		"pandora_sentry": Param.PandoraSentry,
		"api_prefix":     Param.ApiPrefix,
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
		"buildId":      Param.BuildId,
		"isFallback":   false,
		"gssp":         true,
		"scriptLoader": []string{},
	}
	return props
}

// fetchShareDetail 从源服务器处抓取share info
func fetchShareDetail(shareID string) (retJson map[string]interface{}, err error) {
	url1 := Param.ApiPrefix + "/api/share/" + shareID
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

// postLoginHandler 官方账号密码登陆
func postLoginHandler(c *gin.Context) {
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
}

// autoLoginHandler 在访问自动登陆页面时自动设置cookie
func autoLoginHandler(c *gin.Context) {
	uuid := c.Param("uuid")
	Token, ExpiryTime, err := db.GetTokenAndExpiryTimeByUUID(uuid)
	if err != nil {
		logger.Error("db.GetTokenAndExpiryTimeByUUID failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Unknown UUID",
		})
		return
	}
	if ExpiryTime.Before(time.Now()) {
		c.String(http.StatusFound, "正在自动更新Token，请稍后...")
		_, err := db.UpdateTokenByUUID(uuid)
		if err != nil {
			c.String(http.StatusBadRequest, "\n更新Token失败")
			logger.Error("pandora.GetTokenByRefreshToken failed", zap.Error(err))
			return
		}
		c.String(http.StatusOK, "\n更新Token成功！")
	}

	// 设置cookie
	cookie := &http.Cookie{
		Name:     "access-token",
		Value:    Token,
		Expires:  ExpiryTime,
		Path:     "/",
		Domain:   "",
		Secure:   false,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(c.Writer, cookie)
	c.Redirect(http.StatusFound, "/")
	c.String(http.StatusOK, "若网页并没有跳转，请手动刷新本页...")
}

// userInfoHandler 获取当前用户的信息
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
					"plugin_display_Params",
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

// chatHandler 主入口函数
func chatHandler(ctx *gin.Context) {
	conversationID := ctx.Param("chatID")
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
		"query":        gin.H{},
		"buildId":      Param.BuildId,
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
		"pandora_sentry": Param.PandoraSentry,
		"api_prefix":     Param.ApiPrefix,
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
		c.Redirect(http.StatusFound, next)
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

func error404(c *gin.Context) {
	props := gin.H{
		"props": gin.H{
			"pageProps": gin.H{"statusCode": 404},
		},
		"page":         "/_error",
		"query":        gin.H{},
		"buildId":      Param.BuildId,
		"nextExport":   true,
		"isFallback":   false,
		"gip":          true,
		"scriptLoader": "[]",
	}
	c.HTML(http.StatusNotFound, "404.html", gin.H{
		"props":          props,
		"pandora_sentry": Param.PandoraSentry,
		"api_prefix":     Param.ApiPrefix,
	})
}
