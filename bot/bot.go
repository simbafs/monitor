package bot

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type UpdateHandleFunc func(*Bot, tgbotapi.Update)

type Cmd struct {
	Handler     UpdateHandleFunc
	Description string
	Hide        bool
}

type Wait struct {
	Key  string
	Hook UpdateHandleFunc
}

type Subscriber struct {
	Value map[string]interface{}
}

type Bot struct {
	Bot         *tgbotapi.BotAPI
	subscribers map[int64]Subscriber
	cmd         map[string]Cmd
	button      map[string]UpdateHandleFunc
	wait        map[int64]Wait
}

func New(bot *tgbotapi.BotAPI, err error) (*Bot, error) {
	if err != nil {
		return nil, err
	}
	b := &Bot{
		Bot:         bot,
		subscribers: map[int64]Subscriber{},
		cmd:         map[string]Cmd{},
		button:      map[string]UpdateHandleFunc{},
		wait:        map[int64]Wait{},
	}

	b.Subscribe(901756183) // @simbafs, for testing

	return b, nil
}

func (b *Bot) HandleUpdates() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.Bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			if update.Message.IsCommand() {
				go b.HandleCmds(update)
			} else if update.Message.Text != "" {
				if b.MatchWatting(update) {
					// TODO

					continue
				}

				log.Printf("unknown message: %s\n", update.Message.Text)
			} else {
				log.Printf("unknown message: %v\n", update.Message)
			}
		} else if update.CallbackQuery != nil {
			go b.HandleButton(update)
		} else {
			log.Printf("unknown update: %v", update)
		}
	}
}
