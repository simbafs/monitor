package main

import (
	"bytes"
	"fmt"
	"log"
	"math"
	mybot "monitor/bot"
	cfg "monitor/config"
	"monitor/history"
	"os"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

var telegramBotToken = os.Getenv("TG_BOT_TOKEN")

var config = cfg.New().Float64("cpu_threshold", "CPU threshold", 75.0).
	Float64("mem_threshold", "Memory threshold", 85.0).
	Float64("increase_threshold", "Increase threshold (in how many standard deviation)", 2.0).
	Int("interval", "Interval", 1)

var (
	cpuUsageHistory = history.New(30*time.Minute, "CPU")    // History of CPU usage percentages
	memUsageHistory = history.New(30*time.Minute, "Memory") // History of memory usage percentages
)

var avgInterval = 10 * time.Minute

func main() {
	bot, err := mybot.New(tgbotapi.NewBotAPI(telegramBotToken))
	if err != nil {
		log.Fatal(err)
	}

	registerCmdsAndBtn(bot)

	// bot.Debug = true

	bot.Boradcast("Bot started")
	fmt.Println("Bot started")

	go bot.HandleUpdates()

	for {
		checkAndNotify(bot)
		time.Sleep(time.Duration(config.GetInt("interval")) * time.Minute) // Check every 5 minutes
	}
}

func registerCmdsAndBtn(bot *mybot.Bot) {
	bot.AddCmd("subscribe", "Subscribe notifications", false, func(b *mybot.Bot, u tgbotapi.Update) {
		if b.IsSubscribed(u.Message.Chat.ID) {
			b.SendMsg(u.Message.Chat.ID, "Already subscribed")
		} else {
			b.Subscribe(u.Message.Chat.ID)
			b.SendMsg(u.Message.Chat.ID, "Subscribed to notifications")
		}
	})

	bot.AddCmd("unsubscribe", "Unsubscribe notifications", false, func(b *mybot.Bot, u tgbotapi.Update) {
		if b.IsSubscribed(u.Message.Chat.ID) {
			b.Unsubscribe(u.Message.Chat.ID)
			b.SendMsg(u.Message.Chat.ID, "Unsubscribed from notifications")
		} else {
			b.SendMsg(u.Message.Chat.ID, "Not subscribed")
		}
	})

	bot.AddCmd("status", "Get server status", false, func(b *mybot.Bot, u tgbotapi.Update) {
		chatID := u.Message.Chat.ID
		cpuPercent, err := cpu.Percent(time.Second, false)
		if err != nil {
			bot.SendMsg(chatID, "Error getting CPU usage")
			return
		}
		memPercent, err := mem.VirtualMemory()
		if err != nil {
			bot.SendMsg(chatID, "Error getting memory usage")
			return
		}

		cpuAvg := cpuUsageHistory.Average(avgInterval)
		memAvg := memUsageHistory.Average(avgInterval)
		cpuStddev := cpuUsageHistory.StdDev(avgInterval)
		memStddev := memUsageHistory.StdDev(avgInterval)

		bot.SendMsg(chatID, fmt.Sprintf(`===Current Value===
CPU: %.2f%%
Memory: %.2f%%

=====Average=====
CPU: %.2f%% (±%.2f)
Memory: %.2f%% (±%.2f)`,
			cpuPercent[0],
			memPercent.UsedPercent,
			cpuAvg, cpuStddev,
			memAvg, memStddev,
		))
	})

	bot.AddCmd("set", "Set config value", false, config.CmdSet)

	bot.AddCmd("config", "Get all config values", false, func(b *mybot.Bot, u tgbotapi.Update) {
		b.SendMsg(u.Message.Chat.ID, config.All())
	})

	bot.AddCmd("plot", "Plot resource usage", false, func(b *mybot.Bot, u tgbotapi.Update) {
		plot(b, u.Message.Chat.ID)
	})

	bot.AddCmd("add", "Manualy add data point (for debug)", true, func(b *mybot.Bot, u tgbotapi.Update) {
		seg := strings.Split(u.Message.Text, " ")
		n := 1
		if len(seg) > 1 {
			var err error
			n, err = strconv.Atoi(seg[1])
			if err != nil {
				b.SendMsg(u.Message.Chat.ID, "Invalid argument")
				return
			}
		}
		for i := 0; i < n; i++ {
			checkAndNotify(bot)
			b.SendMsg(u.Message.Chat.ID, fmt.Sprintf("add %d", i))
			time.Sleep(1 * time.Second)
		}

		b.SendMsg(u.Message.Chat.ID, "Done")
	})

	plotBtn := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Plot", "plot"),
		),
	)

	bot.AddCmd("plotbtn", "Send a message with a button to exec plot command", true, func(b *mybot.Bot, u tgbotapi.Update) {
		// unpin all
		unpinConfig := tgbotapi.UnpinAllChatMessagesConfig{
			ChatID: u.Message.Chat.ID,
		}
		bot.Request(unpinConfig)

		msg := tgbotapi.NewMessage(u.Message.Chat.ID, "Click to plot")
		msg.ReplyMarkup = plotBtn
		m, err := b.Send(msg)

		if err == nil {
			// pin the message
			pin := tgbotapi.PinChatMessageConfig{
				ChatID:              u.Message.Chat.ID,
				MessageID:           m.MessageID,
				DisableNotification: true,
			}
			b.Request(pin)
		}
	})

	bot.AddButton("plot", func(b *mybot.Bot, u tgbotapi.Update) {
		plot(b, u.CallbackQuery.Message.Chat.ID)
	})

	bot.AddCmd("hi", "Example command for waiting", true, func(b *mybot.Bot, u tgbotapi.Update) {
		forceReply := tgbotapi.ForceReply{
			ForceReply:            true,
			InputFieldPlaceholder: "What's your name?",
		}
		msg := tgbotapi.NewMessage(u.Message.Chat.ID, "What's your name?")
		msg.ReplyMarkup = forceReply
		b.Send(msg)
		b.Wait(u.Message.Chat.ID, "name", func(b *mybot.Bot, u tgbotapi.Update) {
			b.SendMsg(u.Message.Chat.ID, "Hello, "+u.Message.Text)
		})
	})

	bot.AddCmd("cancel", "Cancel waiting", true, func(b *mybot.Bot, u tgbotapi.Update) {
		b.Cancel(u.Message.Chat.ID)
		b.SendMsg(u.Message.Chat.ID, "Cancelled")
	})

	bot.AddCmd("menu", "set commands menu", true, setMenu)

	bot.AddCmd("history", "Show history", false, func(b *mybot.Bot, u tgbotapi.Update) {
		seg := strings.Split(u.Message.Text, " ")
		if len(seg) < 2 {
			b.SendMsg(u.Message.Chat.ID, "Invalid argument")
			return
		}

		historyName := seg[1]
		var h *history.History
		switch historyName {
		case "cpu":
			h = cpuUsageHistory
		case "mem":
			h = memUsageHistory

		default:
			b.SendMsg(u.Message.Chat.ID, "Invalid argument")
			return
		}

		b.SendMsg(u.Message.Chat.ID, h.String())
	})
}

// setMenu set the bot's command menu
func setMenu(b *mybot.Bot, u tgbotapi.Update) {
	deleteCommandConfig := tgbotapi.NewDeleteMyCommands()
	b.Request(deleteCommandConfig)

	setConfig := tgbotapi.NewSetMyCommandsWithScope(tgbotapi.NewBotCommandScopeAllPrivateChats())

	for cmd, C := range b.Cmds() {
		if C.Hide {
			continue
		}
		log.Println(cmd, C.Description)
		setConfig.Commands = append(setConfig.Commands, tgbotapi.BotCommand{
			Command:     cmd,
			Description: C.Description,
		})
	}

	log.Println(setConfig.Commands)

	_, err := b.Request(setConfig)
	if err != nil {
		b.SendMsg(u.Message.Chat.ID, err.Error())
	} else {
		b.SendMsg(u.Message.Chat.ID, "doen")
	}
}

// plot plots the CPU and memory usage history and sends the plot to the chat
func plot(b *mybot.Bot, chatID int64) {
	img, err := history.Plot(cpuUsageHistory, memUsageHistory)
	if err != nil {
		b.SendMsg(chatID, "Error plotting")
	}

	var imgBuf bytes.Buffer
	if _, err := img.WriteTo(&imgBuf); err != nil {
		b.SendMsg(chatID, "Error plotting")
	}

	file := tgbotapi.FileBytes{Name: "usage.png", Bytes: imgBuf.Bytes()}

	photo := tgbotapi.NewPhoto(chatID, file)

	if _, err := b.Send(photo); err != nil {
		b.SendMsg(chatID, err.Error())
	}
}

func isSuddenlyIncrease(curr float64, history *history.History) (float64, bool) {
	avg := history.Average(avgInterval)
	stddev := history.StdDev(avgInterval)

	if curr < 5.0 {
		return 0, false
	}

	z := math.Abs((curr - avg) / stddev)

	threshold := config.GetFloat64("increase_threshold")

	return z, z > threshold
}

func checkAndNotify(bot *mybot.Bot) {
	cpuPercent, err := cpu.Percent(time.Second, false)

	if err == nil {
		if cpuPercent[0] > config.GetFloat64("cpu_threshold") {
			bot.Boradcast(fmt.Sprintf("High CPU usage detected: %.2f%%", cpuPercent[0]))
		}

		if z, yes := isSuddenlyIncrease(cpuPercent[0], cpuUsageHistory); yes {
			bot.Boradcast(fmt.Sprintf("Sudden increase in CPU usage detected: %.2f%% (z = %.2f)", cpuPercent[0], z))
		}

		cpuUsageHistory.Append(cpuPercent[0])
	}

	memStat, err := mem.VirtualMemory()
	if err == nil {
		if memStat.UsedPercent > config.GetFloat64("mem_threshold") {
			bot.Boradcast(fmt.Sprintf("High memory usage detected: %.2f%%", memStat.UsedPercent))
		}

		if z, yes := isSuddenlyIncrease(memStat.UsedPercent, memUsageHistory); yes {
			bot.Boradcast(fmt.Sprintf("Sudden increase in memory usage detected: %.2f%% (z = %.2f)", memStat.UsedPercent, z))
		}

		memUsageHistory.Append(memStat.UsedPercent)
	}
}
