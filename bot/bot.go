package bot

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type UpdateHandleFunc func(*Bot, tgbotapi.Update)

type Cmd struct {
	Handler     UpdateHandleFunc
	Description string
}

type Bot struct {
	Bot         *tgbotapi.BotAPI
	subscribers map[int64]context.Context
	cmd         map[string]Cmd
	button      map[string]UpdateHandleFunc
}

func New(bot *tgbotapi.BotAPI, err error) (*Bot, error) {
	if err != nil {
		return nil, err
	}
	return &Bot{
		Bot: bot,
		subscribers: map[int64]context.Context{
			901756183: context.Background(), // @simbafs
		},
		cmd:    map[string]Cmd{},
		button: map[string]UpdateHandleFunc{},
	}, nil
}

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

// Context returns the context of a user.
func (b *Bot) Context(chatID int64) context.Context {
	return b.subscribers[chatID]
}
