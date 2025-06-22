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
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", s.token)
	body := map[string]interface{}{
		"chat_id": userID,
		"text":    text,
	}
	bodyBytes, _ := json.Marshal(body)

	resp, err := s.client.Post(url, "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("telegram api error on send message (%d): %s", resp.StatusCode, string(respBody))
	}
	return nil
}
