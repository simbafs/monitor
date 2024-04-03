package bot

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type CmdFunc func(*Bot, tgbotapi.Update)

type Cmd struct {
	Cmd         CmdFunc
	Description string
}

type Bot struct {
	Bot         *tgbotapi.BotAPI
	subscribers map[int64]context.Context
	cmd         map[string]Cmd
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
		cmd: map[string]Cmd{},
	}, nil
}

// subscription management

// Subscribe adds a user to the list of subscribers.
func (b *Bot) Subscribe(chatID int64) {
	b.subscribers[chatID] = context.Background()
}

// UnSubscribe removes a user from the list of subscribers.
func (b *Bot) Unsubscribe(chatID int64) {
	delete(b.subscribers, chatID)
}

// IsSubscribed checks if a user is subscribed.
func (b *Bot) IsSubscribed(chatID int64) bool {
	_, ok := b.subscribers[chatID]
	return ok
}

// N returns the number of subscribers.
func (b *Bot) N() int {
	return len(b.subscribers)
}

// Message

// Send sends a Chattable to a user.
func (b *Bot) Send(msg tgbotapi.Chattable) {
	b.Bot.Send(msg)
}

// SendMsg sends a message to a user.
func (b *Bot) SendMsg(chatID int64, msg string) {
	b.Send(tgbotapi.NewMessage(chatID, msg))
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

// command

func (b *Bot) AddCmd(cmd string, description string, f CmdFunc) {
	b.cmd[cmd] = Cmd{
		Cmd:         f,
		Description: description,
	}
}

func (b *Bot) HandleCmds(update tgbotapi.Update) {
	if update.Message == nil || !update.Message.IsCommand() {
		return
	}

	cmd, ok := b.cmd[update.Message.Command()]
	if !ok {
		b.Help(update)
	} else {
		cmd.Cmd(b, update)
	}
}

func (b *Bot) Help(update tgbotapi.Update) {
	res := "Available commands:\n/help - Show this message\n"
	for cmd := range b.cmd {
		res += "/" + cmd + " - " + b.cmd[cmd].Description + "\n"
	}

	b.SendMsg(update.Message.Chat.ID, res)
}
