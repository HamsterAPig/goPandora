package pandora

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	logger "goPandora/internal/log"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
)

const userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/113.0.0.0 Safari/537.36"

func Auth0(userName string, password string, mfaCode string, proxy string) (string, string, error) {
	// 正则表达式模式用于验证电子邮件地址
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	// 使用正则表达式验证电子邮件地址格式
	match, _ := regexp.MatchString(pattern, userName)
	if !match {
		return "", "", fmt.Errorf("%s is not a valid email address", userName)
	}
	client := http.Client{
		Jar: createCookieJar(),
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// 禁用跟随301跳转
			return http.ErrUseLastResponse
		},
	}

	codeVerifier, _ := GenerateCodeVerifier()
	codeChallenge := GenerateCodeChallenge(codeVerifier)

	// 获取State
	url1 := "https://auth0.openai.com/authorize?client_id=pdlLIX2Y72MIl2rhLhTE9VV9bN905kBh&audience=https%3A%2F%2Fapi.openai.com%2Fv1&redirect_uri=com.openai.chat%3A%2F%2Fauth0.openai.com%2Fios%2Fcom.openai.chat%2Fcallback&scope=openid%20email%20profile%20offline_access%20model.request%20model.read%20organization.read%20offline&response_type=code&code_challenge=HlLnX9QkMGL0gGRBoyjtXtWcuIc9_t_CTNyNX8dLahk&code_challenge_method=S256&prompt=login"
	url1 = strings.Replace(url1, "code_challenge=HlLnX9QkMGL0gGRBoyjtXtWcuIc9_t_CTNyNX8dLahk", "code_challenge="+codeChallenge, 1)
	req1, err := http.NewRequest(http.MethodGet, url1, nil)
	if err != nil {
		return "", "", fmt.Errorf("create request_1 error: %s", err)
	}
	req1.Header.Set("Referer", "https://ios.chat.openai.com/")
	req1.Header.Set("User-Agent", userAgent)
	resp1, err := client.Do(req1)
	if err != nil {
		return "", "", fmt.Errorf("do request_1 error: %s", err)
	}
	defer resp1.Body.Close()
	if resp1.StatusCode != http.StatusFound {
		return "", "", fmt.Errorf("request_1 rate limit hit")
	}
	location := resp1.Header.Get("Location")
	parsedURL, err := url.Parse(location)
	if err != nil {
		return "", "", fmt.Errorf("parse location error: %s", err)
	}
	queryParams := parsedURL.Query()
	state := queryParams.Get("state")
	logger.Debug("state", zap.String("state", state))

	// POST 用户名数据
	// 构建请求体数据
	formData := url.Values{}
	formData.Set("state", state)
	formData.Set("username", userName)
	formData.Set("js-available", "true")
	formData.Set("webauthn-available", "true")
	formData.Set("is-brave", "false")
	formData.Set("webauthn-platform-available", "false")
	formData.Set("action", "default")
	body := strings.NewReader(formData.Encode())

	url2 := "https://auth0.openai.com/u/login/identifier?state=" + state
	req2, err := http.NewRequest(http.MethodPost, url2, body)
	if err != nil {
		return "", "", fmt.Errorf("create request_2 error: %s", err)
	}
	req2.Header.Set("User-Agent", userAgent)
	req2.Header.Set("Referer", url2)
	req2.Header.Set("Origin", "https://auth0.openai.com")
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp2, err := client.Do(req2)
	if err != nil {
		return "", "", fmt.Errorf("do request_2 error: %s", err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusFound {
		return "", "", fmt.Errorf("request_2 Error check email")
	}

	// POST用户名与密码
	formData = url.Values{}
	formData.Set("state", state)
	formData.Set("username", userName)
	formData.Set("password", password)
	formData.Set("action", "default")
	body = strings.NewReader(formData.Encode())
	url3 := "https://auth0.openai.com/u/login/password?state=" + state
	req3, err := http.NewRequest(http.MethodPost, url3, body)
	if err != nil {
		return "", "", fmt.Errorf("create request_3 error: %s", err)
	}
	req3.Header.Set("User-Agent", userAgent)
	req3.Header.Set("Origin", "https://auth0.openai.com")
	req3.Header.Set("Referer", url3)
	req3.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp3, err := client.Do(req3)
	if err != nil {
		return "", "", fmt.Errorf("do request_3 error: %s", err)
	}
	defer resp3.Body.Close()
	if resp3.StatusCode != http.StatusFound {
		return "", "", fmt.Errorf("request_3 Error")
	} else if resp3.StatusCode == http.StatusBadRequest {
		return "", "", fmt.Errorf("request_3 Wrong email or password")
	}
	location = resp3.Header.Get("Location")
	parsedURL, err = url.Parse(location)
	if err != nil {
		return "", "", fmt.Errorf("parse location error: %s", err)
	}
	queryParams = parsedURL.Query()
	state1 := queryParams.Get("state")
	logger.Debug("state", zap.String("state_1", state1))

	// 获取Callback
	url4 := "https://auth0.openai.com/authorize/resume?state=" + state1 + "&Referer=https://auth0.openai.com/u/login/password?state=" + state
	req4, err := http.NewRequest(http.MethodGet, url4, nil)
	req4.Header.Set("User-Agent", userAgent)
	resp4, err := client.Do(req4)
	if err != nil {
		return "", "", fmt.Errorf("do request_4 error: %s", err)
	}
	defer resp4.Body.Close()
	if resp4.StatusCode != http.StatusFound {
		return "", "", fmt.Errorf("request_4 Error")
	}
	location = resp4.Header.Get("Location")
	logger.Debug("location", zap.String("location", location))
	parsedURL, err = url.Parse(location)
	if err != nil {
		return "", "", fmt.Errorf("parse location error: %s", err)
	}
	queryParams = parsedURL.Query()
	code := queryParams.Get("code")
	logger.Debug("code", zap.String("code", code))
	accessToken, refreshToken, err := GetTokenAndRefreshTokenByCode(code, codeVerifier)
	logger.Debug("accessToken", zap.String("accessToken", accessToken))
	logger.Debug("refreshToken", zap.String("refreshToken", refreshToken))
	return accessToken, refreshToken, nil
}

// GetTokenAndRefreshTokenByCode 通过code与codeVerifier获取token与refresh token
func GetTokenAndRefreshTokenByCode(code string, codeVerifier string) (string, string, error) {
	client := http.Client{}
	url5 := "https://auth0.openai.com/oauth/token"
	formData := url.Values{}
	formData.Set("redirect_uri", "com.openai.chat://auth0.openai.com/ios/com.openai.chat/callback")
	formData.Set("grant_type", "authorization_code")
	formData.Set("client_id", "pdlLIX2Y72MIl2rhLhTE9VV9bN905kBh")
	formData.Set("code", code)
	formData.Set("code_verifier", codeVerifier)
	req5, err := http.NewRequest(http.MethodPost, url5, strings.NewReader(formData.Encode()))
	if err != nil {
		return "", "", fmt.Errorf("create request_5 error: %s", err)
	}
	req5.Header.Set("User-Agent", userAgent)
	req5.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req5)
	if err != nil {
		return "", "", fmt.Errorf("do request_5 error: %s", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("can't get token")
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("read body error: %s", err)
	}
	jsonStr := string(body)
	var data map[string]interface{}
	err = json.Unmarshal([]byte(jsonStr), &data)
	return data["access_token"].(string), data["refresh_token"].(string), nil
}

// createCookieJar 创建持久化cookie的jar
func createCookieJar() *cookiejar.Jar {
	jar, _ := cookiejar.New(nil)
	return jar
}
