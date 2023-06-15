package db

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"goPandora/config"
	logger "goPandora/internal/log"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type ShareToken struct {
	ID            uint `gorm:"primary_key:autoIncrement"`
	UserID        string
	UniqueName    string
	ExpiresTime   int64
	ExpiresTimeAt time.Time
	SiteLimit     string
	SK            string    `gorm:"unique"`
	UpdateTime    time.Time `gorm:"autoUpdateTime"`
	Comment       string
}

type faseOpenShareToken struct {
	ExpireAt          int64  `json:"expire_at"`
	ShowConversations bool   `json:"show_conversations"`
	ShowUserinfo      bool   `json:"show_userinfo"`
	SiteLimit         string `json:"site_limit"`
	TokenKey          string `json:"token_key"`
	UniqueName        string `json:"unique_name"`
}

const userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/113.0.0.0 Safari/537.36"

func CreateShareToken(userID string, uniqueName string, ExpireTime time.Duration, SiteLimit string, comment string) error {
	logger.Info("CreateShareToken", zap.String("userID", userID), zap.String("uniqueName", uniqueName), zap.Duration("ExpireTime", ExpireTime), zap.String("SiteLimit", SiteLimit), zap.String("comment", comment))
	shareTokenValue := ShareToken{
		UserID:     userID,
		UniqueName: uniqueName,
		SiteLimit:  SiteLimit,
		Comment:    comment,
	}
	var user User
	res := db.Where("user_id = ?", userID).First(&user)
	if res.RowsAffected == 0 {
		return fmt.Errorf("RecordNotFound")
	}

	err := getShareToken(&shareTokenValue, user.Token, ExpireTime)
	if err != nil {
		return fmt.Errorf("getShareToken failed: %w", err)
	}

	res = db.Save(&shareTokenValue)
	return res.Error
}

func getShareToken(shareTokenStruct *ShareToken, token string, ExpireTime time.Duration) error {
	logger.Info("getShareToken", zap.String("token", token))
	target := "https://ai.fakeopen.com/token/register"
	var proxyURL *url.URL
	var err error
	if len(config.Conf.MainConfig.ProxyGroup) > 0 {
		proxyURL, err = url.Parse(config.Conf.MainConfig.ProxyGroup[0])
		if err != nil {
			return fmt.Errorf("url.Parse failed: %w", err)
		}
	}
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
	}

	dataFrom := url.Values{}
	dataFrom.Add("unique_name", shareTokenStruct.UniqueName)
	dataFrom.Add("access_token", token)
	dataFrom.Add("expires_in", fmt.Sprintf("%v", ExpireTime.Seconds()))
	dataFrom.Add("site_limit", shareTokenStruct.SiteLimit)
	body := strings.NewReader(dataFrom.Encode())

	req, err := http.NewRequest(http.MethodPost, target, body)
	if err != nil {
		return fmt.Errorf("http.NewRequest failed: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("client.Do failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("resp.StatusCode != http.StatusOK")
	}
	rawjson, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("io.ReadAll failed: %w", err)
	}
	var data faseOpenShareToken
	err = json.Unmarshal(rawjson, &data)
	if err != nil {
		return fmt.Errorf("json.Unmarshal failed: %w", err)
	}
	shareTokenStruct.SK = data.TokenKey
	shareTokenStruct.UniqueName = data.UniqueName
	shareTokenStruct.ExpiresTimeAt = time.Unix(data.ExpireAt, 0)
	shareTokenStruct.ExpiresTime = int64(ExpireTime * time.Second)
	return nil
}

func UpdateAllShareToken() error {
	var shareTokens []ShareToken
	db.Find(&shareTokens)
	for _, shareToken := range shareTokens {
		_, expiryTime, err := GetTokenAndExpiryTimeByUserID(shareToken.UserID)
		if err != nil {
			return fmt.Errorf("GetTokenAndExpiryTimeByUserID failed: %w", err)
		}
		if expiryTime.Before(time.Now()) {
			_, err = UpdateTokenByUserID(shareToken.UserID)
			if err != nil {
				return fmt.Errorf("UpdateTokenByUserID failed: %w", err)
			}
			err = CreateShareToken(shareToken.UserID, shareToken.UniqueName, time.Duration(shareToken.ExpiresTime)*time.Second, shareToken.SiteLimit, shareToken.Comment)
			if err != nil {
				return fmt.Errorf("update Share Token failed: %w", err)
			}
		} else {
			logger.Info("Token is not expired, not update", zap.String("share token", shareToken.SK))
		}
	}
	return nil
}

func ListAllShareToken() {
	var shareTokens []ShareToken
	db.Find(&shareTokens)
	for _, shareToken := range shareTokens {
		fmt.Printf("%+v\n", shareToken)
	}
}
