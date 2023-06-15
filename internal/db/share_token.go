package db

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type ShareToken struct {
	ID          uint `gorm:"primary_key:autoIncrement"`
	UserID      string
	UniqueName  string
	ExpiresTime int64
	SiteLimit   string
	SK          string    `gorm:"unique"`
	UpdateTime  time.Time `gorm:"autoUpdateTime"`
	Comment     string
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
	return nil
}

func getShareToken(shareTokenStruct *ShareToken, token string, ExpireTime time.Duration) error {
	target := "https://ai.fakeopen.com/token/register"
	client := &http.Client{}

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
	shareTokenStruct.ExpiresTime = data.ExpireAt
	return nil
}

func GetAllShareToken() []ShareToken {
	var shareTokens []ShareToken
	db.Find(&shareTokens)
	return shareTokens
}
