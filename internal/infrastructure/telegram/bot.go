package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"tribute-back/internal/config"
)

// BotService handles interactions with the Telegram Bot API.
type BotService struct {
	token       string
	client      *http.Client
	adminChatID string
}

// NewBotService creates a new instance of the BotService.
func NewBotService() (*BotService, error) {
	token := config.GetEnv("TELEGRAM_BOT_TOKEN", "")
	if token == "" {
		return nil, fmt.Errorf("TELEGRAM_BOT_TOKEN environment variable not set")
	}
	adminChatID := config.GetEnv("TELEGRAM_ADMIN_CHAT_ID", "")
	if adminChatID == "" {
		return nil, fmt.Errorf("TELEGRAM_ADMIN_CHAT_ID environment variable not set")
	}

	return &BotService{
		token:       token,
		client:      &http.Client{},
		adminChatID: adminChatID,
	}, nil
}

// InlineKeyboardButton represents a single button in an inline keyboard.
type InlineKeyboardButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data"`
}

// InlineKeyboardMarkup represents an inline keyboard.
type InlineKeyboardMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}

// sendPhoto sends a photo to a specific chat.
func (s *BotService) sendPhoto(chatID string, photo io.Reader, caption string) error {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	fw, err := w.CreateFormFile("photo", "passport.jpg")
	if err != nil {
		return err
	}
	if _, err = io.Copy(fw, photo); err != nil {
		return err
	}
	w.WriteField("chat_id", chatID)
	w.WriteField("caption", caption)
	w.Close()

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendPhoto", s.token)
	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("telegram api error (%d): %s", resp.StatusCode, string(body))
	}
	return nil
}

// SendVerificationRequest sends the user's documents to the admin chat with action buttons.
func (s *BotService) SendVerificationRequest(userID int64, userPhoto io.Reader, userPassport io.Reader) error {
	if err := s.sendPhoto(s.adminChatID, userPhoto, fmt.Sprintf("User Photo for UserID: %d", userID)); err != nil {
		return fmt.Errorf("failed to send user photo: %w", err)
	}
	if err := s.sendPhoto(s.adminChatID, userPassport, fmt.Sprintf("User Passport for UserID: %d", userID)); err != nil {
		return fmt.Errorf("failed to send user passport: %w", err)
	}

	// Send the message with inline keyboard for actions
	text := fmt.Sprintf("Please verify user with ID: %d", userID)
	callbackApprove := fmt.Sprintf("verify_approve_%d", userID)
	callbackReject := fmt.Sprintf("verify_reject_%d", userID)
	keyboard := InlineKeyboardMarkup{
		InlineKeyboard: [][]InlineKeyboardButton{
			{
				{Text: "Подтвердить", CallbackData: callbackApprove},
				{Text: "Отклонить", CallbackData: callbackReject},
			},
		},
	}
	keyboardBytes, err := json.Marshal(keyboard)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", s.token)
	body := map[string]interface{}{
		"chat_id":      s.adminChatID,
		"text":         text,
		"reply_markup": json.RawMessage(keyboardBytes),
	}
	bodyBytes, _ := json.Marshal(body)

	resp, err := s.client.Post(url, "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("telegram api error on send keyboard (%d): %s", resp.StatusCode, string(respBody))
	}
	return nil
}

// DeleteMessage deletes a message from a chat.
func (s *BotService) DeleteMessage(chatID int64, messageID int) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/deleteMessage", s.token)
	body := map[string]interface{}{
		"chat_id":    chatID,
		"message_id": messageID,
	}
	bodyBytes, _ := json.Marshal(body)

	resp, err := s.client.Post(url, "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// We don't strictly need to check for error, as the message might be old
	return nil
}

// SendMessage sends a simple text message to a user.
func (s *BotService) SendMessage(userID int64, text string) error {
	fmt.Printf("Sending message to user %d: %s\n", userID, text)
	fmt.Printf("Using bot token: %s...\n", s.token[:10]) // Show first 10 chars for debugging

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", s.token)
	body := map[string]interface{}{
		"chat_id": userID,
		"text":    text,
	}
	bodyBytes, _ := json.Marshal(body)

	fmt.Printf("Sending request to: %s\n", url)
	fmt.Printf("Request body: %s\n", string(bodyBytes))

	resp, err := s.client.Post(url, "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		fmt.Printf("HTTP request failed: %v\n", err)
		return err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	fmt.Printf("Response status: %d\n", resp.StatusCode)
	fmt.Printf("Response body: %s\n", string(respBody))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram api error on send message (%d): %s", resp.StatusCode, string(respBody))
	}

	fmt.Printf("Message sent successfully to user %d\n", userID)
	return nil
}

// ChatMember represents a member in a chat
type ChatMember struct {
	Status string `json:"status"`
	User   struct {
		ID       int64  `json:"id"`
		Username string `json:"username"`
	} `json:"user"`
}

// TelegramResponse represents a generic Telegram API response
type TelegramResponse struct {
	OK     bool        `json:"ok"`
	Result interface{} `json:"result"`
}

// CheckChannelMembership checks if a user is a member of a channel and their role
func (s *BotService) CheckChannelMembership(channelUsername string, userID int64) (*ChatMember, error) {
	// Remove @ if present
	if len(channelUsername) > 0 && channelUsername[0] == '@' {
		channelUsername = channelUsername[1:]
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/getChatMember", s.token)
	body := map[string]interface{}{
		"chat_id": "@" + channelUsername,
		"user_id": userID,
	}
	bodyBytes, _ := json.Marshal(body)

	resp, err := s.client.Post(url, "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to check channel membership: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("telegram api error on getChatMember (%d): %s", resp.StatusCode, string(respBody))
	}

	var response TelegramResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !response.OK {
		return nil, fmt.Errorf("telegram api returned error: %v", response)
	}

	// Parse the result as ChatMember
	resultBytes, _ := json.Marshal(response.Result)
	var chatMember ChatMember
	if err := json.Unmarshal(resultBytes, &chatMember); err != nil {
		return nil, fmt.Errorf("failed to parse chat member: %w", err)
	}

	return &chatMember, nil
}
