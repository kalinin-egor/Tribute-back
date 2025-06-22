package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
	"tribute-back/internal/config"
)

// InitDataUser represents the user part of the initData.
type InitDataUser struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name,omitempty"`
	Username  string `json:"username,omitempty"`
}

// ParsedInitData holds the structured data from the initData string.
type ParsedInitData struct {
	User     InitDataUser `json:"user"`
	AuthDate int64        `json:"auth_date"`
	Hash     string       `json:"hash"`
}

// TelegramAuthService provides methods to validate Telegram initData.
type TelegramAuthService struct {
	botToken string
}

// NewTelegramAuthService creates a new instance of the service.
func NewTelegramAuthService() (*TelegramAuthService, error) {
	token := config.GetEnv("TELEGRAM_BOT_TOKEN", "")
	if token == "" {
		return nil, fmt.Errorf("TELEGRAM_BOT_TOKEN environment variable not set")
	}
	return &TelegramAuthService{botToken: token}, nil
}

// Validate validates the initData string against the bot token.
// It returns the parsed user data if valid, or an error otherwise.
func (s *TelegramAuthService) Validate(initData string) (*ParsedInitData, error) {
	q, err := url.ParseQuery(initData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse initData query: %w", err)
	}

	hash := q.Get("hash")
	if hash == "" {
		return nil, fmt.Errorf("hash field is missing from initData")
	}

	var dataCheckPairs []string
	for k, v := range q {
		if k != "hash" {
			dataCheckPairs = append(dataCheckPairs, fmt.Sprintf("%s=%s", k, v[0]))
		}
	}
	sort.Strings(dataCheckPairs)
	dataCheckString := strings.Join(dataCheckPairs, "\n")

	secretKey := hmac.New(sha256.New, []byte("WebAppData"))
	secretKey.Write([]byte(s.botToken))

	hmacHash := hmac.New(sha256.New, secretKey.Sum(nil))
	hmacHash.Write([]byte(dataCheckString))
	calculatedHash := hex.EncodeToString(hmacHash.Sum(nil))

	if calculatedHash != hash {
		return nil, fmt.Errorf("hash validation failed")
	}

	var parsedData ParsedInitData
	userJSON := q.Get("user")
	if userJSON == "" {
		return nil, fmt.Errorf("user field is missing from initData")
	}
	if err := json.Unmarshal([]byte(userJSON), &parsedData.User); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user data: %w", err)
	}

	authDate, _ := strconv.ParseInt(q.Get("auth_date"), 10, 64)
	parsedData.AuthDate = authDate

	// Optional: Check if the data is recent
	if time.Since(time.Unix(parsedData.AuthDate, 0)) > time.Hour*24 {
		return nil, fmt.Errorf("initData is outdated")
	}

	return &parsedData, nil
}
