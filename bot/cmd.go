package bot

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

// command

func (b *Bot) AddCmd(cmd string, description string, f UpdateHandleFunc) {
	b.cmd[cmd] = Cmd{
		Handler:     f,
		Description: description,
	}
}

func (b *Bot) HandleCmds(update tgbotapi.Update) {
	cmd, ok := b.cmd[update.Message.Command()]
	if !ok {
		b.Help(update)
	} else {
		cmd.Handler(b, update)
	}
}

func (b *Bot) Help(update tgbotapi.Update) error {
	res := "Available commands:\n/help - Show this message\n"
	for cmd := range b.cmd {
		res += "/" + cmd + " - " + b.cmd[cmd].Description + "\n"
	}

	_, err := b.SendMsg(update.Message.Chat.ID, res)
	return err
}
