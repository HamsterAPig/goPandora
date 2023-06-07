package pandora

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type RefreshData struct {
	RedirectUri  string `json:"redirect_uri"`
	GrantType    string `json:"grant_type"`
	ClientId     string `json:"client_id"`
	RefreshToken string `json:"refresh_token"`
}

// refreshPostToken 向网页post数据
func refreshPostToken(url string, data RefreshData, userAgent string) (*http.Response, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("编码数据失败: %v", err)
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent)
	client := &http.Client{}
	rep, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}

	return rep, nil
}

// GetTokenByRefreshToken 依据refresh_token获取access_token
func GetTokenByRefreshToken(RefreshToken string) (string, error) {
	const userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/113.0.0.0 Safari/537.36"
	data := RefreshData{
		RedirectUri: "com.openai.chat://auth0.openai.com/ios/com.openai.chat/callback",
		GrantType:   "refresh_token",
		ClientId:    "pdlLIX2Y72MIl2rhLhTE9VV9bN905kBh",
	}
	data.RefreshToken = RefreshToken
	url := "https://auth0.openai.com/oauth/token"
	rep, err := refreshPostToken(url, data, userAgent)
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
		}
	}(rep.Body)
	if 200 != rep.StatusCode {
		return "", fmt.Errorf("获取Token失败: %v", rep.StatusCode)
	}
	body, err := io.ReadAll(rep.Body)
	var jsonData map[string]interface{}
	err = json.Unmarshal([]byte(body), &jsonData)
	if err != nil {
		return "", fmt.Errorf("解析JSON失败: %v", err)
	}
	accessToken := jsonData["access_token"].(string)
	return accessToken, nil
}
