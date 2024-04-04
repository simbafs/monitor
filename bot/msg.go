package bot

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

// Message

func (b *Bot) Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error) {
	return b.Bot.Request(c)
}

// Send sends a Chattable to a user.
func (b *Bot) Send(msg tgbotapi.Chattable) (tgbotapi.Message, error) {
	return b.Bot.Send(msg)
}

// SendMsg sends a message to a user.
func (b *Bot) SendMsg(chatID int64, msg string) (tgbotapi.Message, error) {
	return b.Send(tgbotapi.NewMessage(chatID, msg))
}

// Broadcast sends a message to all subscribers.
func (b *Bot) Boradcast(msg string) {
	for chatID := range b.subscribers {
		b.SendMsg(chatID, msg)
	}
}
