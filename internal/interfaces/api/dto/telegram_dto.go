package dto

// TelegramUpdate represents the incoming webhook payload from Telegram.
type TelegramUpdate struct {
	UpdateID      int            `json:"update_id"`
	CallbackQuery *CallbackQuery `json:"callback_query"`
}

// CallbackQuery represents the callback query from an inline button press.
type CallbackQuery struct {
	ID      string   `json:"id"`
	From    User     `json:"from"`
	Message *Message `json:"message"`
	Data    string   `json:"data"`
}

// Message represents a Telegram message.
type Message struct {
	MessageID int  `json:"message_id"`
	Chat      Chat `json:"chat"`
}

// Chat represents a conversation.
type Chat struct {
	ID int64 `json:"id"`
}

// User represents a Telegram user or bot.
type User struct {
	ID int64 `json:"id"`
}
