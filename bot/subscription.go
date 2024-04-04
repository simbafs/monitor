package bot

import "context"

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
