package bot

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

// TODO: chain waiting like this
// b.Wait(chatID, "key1").
//   Wait("key2").
//   Wait("key3").
//   Finilly(func(b *Bot, update tgbotapi.Update) {})

func (b *Bot) Wait(chatID int64, key string, hook UpdateHandleFunc) {
	b.wait[chatID] = Wait{
		Key:  key,
		Hook: hook,
	}
}

func (b *Bot) Cancel(chatID int64) {
	delete(b.wait, chatID)
}

func (b *Bot) IsWaiting(chatID int64) bool {
	_, ok := b.wait[chatID]
	return ok
}

func (b *Bot) MatchWatting(update tgbotapi.Update) bool {
	chatID := update.Message.Chat.ID
	if s, ok := b.subscribers[chatID]; ok {
		if b.IsWaiting(chatID) {
			w := b.wait[chatID]
			s.Value[w.Key] = update.Message.Text
			w.Hook(b, update)
			return true
		}
	}
	return false
}
