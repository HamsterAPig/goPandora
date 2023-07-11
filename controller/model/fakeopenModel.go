package model

type ShareTokenResponse struct {
	Email    string `json:"email"`
	ExpireAt int64  `json:"expire_at"`
	UserID   string `json:"user_id"`
}
