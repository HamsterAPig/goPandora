package pandora

import (
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

func Auth0(userName string, password string, mfaCode string, proxy string) (string, error) {
	const userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/113.0.0.0 Safari/537.36"
	// 正则表达式模式用于验证电子邮件地址
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	// 使用正则表达式验证电子邮件地址格式
	match, _ := regexp.MatchString(pattern, userName)
	if !match {
		return "", fmt.Errorf("%s is not a valid email address", userName)
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
		return "", fmt.Errorf("create request_1 error: %s", err)
	}
	req1.Header.Set("Referer", "https://ios.chat.openai.com/")
	req1.Header.Set("User-Agent", userAgent)
	resp1, err := client.Do(req1)
	if err != nil {
		return "", fmt.Errorf("do request_1 error: %s", err)
	}
	defer resp1.Body.Close()
	if resp1.StatusCode != http.StatusFound {
		return "", fmt.Errorf("request_1 rate limit hit")
	}
	location := resp1.Header.Get("Location")
	parsedURL, err := url.Parse(location)
	if err != nil {
		return "", fmt.Errorf("parse location error: %s", err)
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
		return "", fmt.Errorf("create request_2 error: %s", err)
	}
	req2.Header.Set("User-Agent", userAgent)
	req2.Header.Set("Referer", url2)
	req2.Header.Set("Origin", "https://auth0.openai.com")
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp2, err := client.Do(req2)
	if err != nil {
		return "", fmt.Errorf("do request_2 error: %s", err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusFound {
		return "", fmt.Errorf("request_2 Error check email")
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
		return "", fmt.Errorf("create request_3 error: %s", err)
	}
	req3.Header.Set("User-Agent", userAgent)
	req3.Header.Set("Origin", "https://auth0.openai.com")
	req3.Header.Set("Referer", url3)

	resp3, err := client.Do(req3)
	if err != nil {
		return "", fmt.Errorf("do request_3 error: %s", err)
	}
	defer resp3.Body.Close()
	if resp3.StatusCode != http.StatusFound {
		body, _ := io.ReadAll(resp3.Body)
		return string(body), fmt.Errorf("request_3 Error")
	}
	location = resp3.Header.Get("Location")
	parsedURL, err = url.Parse(location)
	if err != nil {
		return "", fmt.Errorf("parse location error: %s", err)
	}
	queryParams = parsedURL.Query()
	state = queryParams.Get("state")
	logger.Debug("state", zap.String("state", state))
	return "", nil
}

// createCookieJar 创建持久化cookie的jar
func createCookieJar() *cookiejar.Jar {
	jar, _ := cookiejar.New(nil)
	return jar
}
