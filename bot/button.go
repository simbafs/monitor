package bot

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

func (b *Bot) AddButton(data string, f UpdateHandleFunc) {
	b.button[data] = f
}

func (b *Bot) HandleButton(update tgbotapi.Update) {
	data := update.CallbackQuery.Data
	f, ok := b.button[data]
	if !ok {
		return
	}
	f(b, update)
}
