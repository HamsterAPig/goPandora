package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"goPandora/config"
	logger "goPandora/internal/log"
	"goPandora/model"
	"io"
	"net/http"
	"time"
)

var validURLs = []string{
	"/sitemap.xml",
	"/robots.txt",
	"/.well-known/security.txt",
}

func NotFoundHandler(c *gin.Context) {
	clientIP := c.ClientIP()
	requestURL := c.Request.URL.String()
	if contains(validURLs, requestURL) {
		c.Redirect(http.StatusFound, "/")
	} else if config.Conf.CloudflareConfig.APIKey != "" {
		go func() {
			targetUrl := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/firewall/access_rules/rules", config.Conf.CloudflareConfig.ZoneID)
			var notes string
			if config.Conf.CloudflareConfig.Notes == "" {
				notes = "blocked ip from goPandora"
			} else {
				notes = config.Conf.CloudflareConfig.Notes
			}
			data := model.CreateCloudflareIPRulesModel{
				Configuration: struct {
					Target string `json:"target"`
					Value  string `json:"value"`
				}{
					Target: "ip",
					Value:  clientIP,
				},
				Mode:  "block",
				Notes: notes + " " + time.Now().Format("2006-01-02 15:04:05") + " " + c.Request.URL.String(),
			}
			jsonData, err := json.Marshal(data)
			if err != nil {
				logger.Error("json.Marshal failed", zap.Error(err))
				c.Abort()
				return
			}
			// 创建请求体
			requestBody := bytes.NewBuffer(jsonData)

			// 创建HTTP请求
			req, err := http.NewRequest(http.MethodPost, targetUrl, requestBody)
			if err != nil {
				fmt.Println("Error creating request:", err)
				return
			}

			// 添加头信息
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Auth-Key", config.Conf.CloudflareConfig.APIKey)
			req.Header.Set("X-Auth-Email", config.Conf.CloudflareConfig.Email)

			// 发送HTTP请求
			client := http.DefaultClient
			response, err := client.Do(req)
			if err != nil {
				logger.Error("client.Do failed", zap.Error(err))
				c.Abort()
				return
			}
			defer response.Body.Close()
			if response.StatusCode != http.StatusOK {
				body, err := io.ReadAll(response.Body)
				if err != nil {
					logger.Error("读取响应内容失败", zap.Error(err))
					c.Abort()
					return
				}

				// 解析JSON数据为map
				var responseData map[string]interface{}
				err = json.Unmarshal(body, &responseData)

				code, ok := responseData["errors"].([]interface{})[0].(map[string]interface{})["code"].(float64)
				if !ok {
					logger.Error("failed to get code", zap.Error(err))
					return
				}
				if int(code) == 10009 {
					logger.Info("ip is blocked, skip it", zap.String("ip", clientIP))
				} else {
					logger.Error("response code not 200", zap.Error(err), zap.Int("cf error code", int(code)), zap.String("ip", clientIP))
				}
				c.Abort()
				return
			}
			logger.Info("Blocked a ip", zap.String("ip", clientIP))
		}()
		c.JSON(http.StatusNotFound, gin.H{
			"error": "not found, and you have been blocked!",
			"uri":   c.Request.URL.String(),
			"ip":    clientIP,
		})
	} else {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "not found",
			"uri":   c.Request.URL.String(),
			"ip":    clientIP,
		})
	}
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}
