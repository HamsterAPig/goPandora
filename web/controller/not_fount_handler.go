package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"goPandora/config"
	logger "goPandora/internal/log"
	"goPandora/web/model"
	"net/http"
)

func NotFoundHandler(c *gin.Context) {
	clientIP := c.ClientIP()
	if config.Conf.CloudflareConfig.APIKey != "" {
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
				Notes: notes,
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
				logger.Error("response.StatusCode != http.StatusOK", zap.Error(err))
				c.Abort()
				return
			}
			logger.Info("Blocked a ip", zap.String("ip", clientIP))
		}()
		c.JSON(http.StatusNotFound, gin.H{
			"error": "not found, and you have been blocked!",
			"ip":    clientIP,
		})
	} else {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "not found",
			"ip":    clientIP,
		})
	}
}
