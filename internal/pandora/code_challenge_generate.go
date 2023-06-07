package pandora

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"strings"
)

// GenerateCodeVerifier 生成一个长度为32的随机code_verifier
func GenerateCodeVerifier() (string, error) {
	// Generate a random token with length 32
	token := make([]byte, 32)
	_, err := rand.Read(token)
	if err != nil {
		return "", err
	}

	// Encode the token in base64url format
	codeVerifier := strings.TrimRight(base64.URLEncoding.EncodeToString(token), "=")
	return codeVerifier, nil
}

// GenerateCodeChallenge 根据code_verifier生成code_challenge
func GenerateCodeChallenge(codeVerifier string) string {
	// Calculate the SHA256 hash of the codeVerifier
	hash := sha256.Sum256([]byte(codeVerifier))

	// Encode the hash in base64url format
	codeChallenge := strings.TrimRight(base64.URLEncoding.EncodeToString(hash[:]), "=")
	return codeChallenge
}
