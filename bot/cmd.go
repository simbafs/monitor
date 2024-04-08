package bot

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

// AddCmd adds a command to the bot
func (b *Bot) AddCmd(cmd string, description string, hide bool, f UpdateHandleFunc) {
	// if cmd contain uppercase, panic
	for _, c := range cmd {
		if c >= 'A' && c <= 'Z' {
			panic("Command must be lowercase")
		}
	}

	b.cmd[cmd] = Cmd{
		Handler:     f,
		Description: description,
		Hide:        hide,
	}
}

// HandleCmds handles commands from the incoming update
func (b *Bot) HandleCmds(update tgbotapi.Update) {
	cmd, ok := b.cmd[update.Message.Command()]
	if !ok {
		b.Help(update)
	} else {
		cmd.Handler(b, update)
	}
}

// Help shows all available commands
func (b *Bot) Help(update tgbotapi.Update) error {
	res := "Available commands:\n/help - Show this message\n"
	for cmd, C := range b.cmd {
		if C.Hide {
			continue
		}
		res += "/" + cmd + " - " + b.cmd[cmd].Description + "\n"
	}

	_, err := b.SendMsg(update.Message.Chat.ID, res)
	return err
}

// Cmds returns all registered commands
func (b *Bot) Cmds() map[string]Cmd {
	return b.cmd
}
