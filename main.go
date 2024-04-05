package main

import (
	"bytes"
	"fmt"
	"log"
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
	Int("interval", "Interval", 1)

var (
	cpuUsageHistory = history.New(30*time.Minute, "CPU")    // History of CPU usage percentages
	memUsageHistory = history.New(30*time.Minute, "Memory") // History of memory usage percentages
)

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
	bot.AddCmd("subscribe", "Subscribe notifications", func(b *mybot.Bot, u tgbotapi.Update) {
		if b.IsSubscribed(u.Message.Chat.ID) {
			b.SendMsg(u.Message.Chat.ID, "Already subscribed")
		} else {
			b.Subscribe(u.Message.Chat.ID)
			b.SendMsg(u.Message.Chat.ID, "Subscribed to notifications")
		}
	})

	bot.AddCmd("unsubscribe", "Unsubscribe notifications", func(b *mybot.Bot, u tgbotapi.Update) {
		if b.IsSubscribed(u.Message.Chat.ID) {
			b.Unsubscribe(u.Message.Chat.ID)
			b.SendMsg(u.Message.Chat.ID, "Unsubscribed from notifications")
		} else {
			b.SendMsg(u.Message.Chat.ID, "Not subscribed")
		}
	})

	bot.AddCmd("status", "Get server status", func(b *mybot.Bot, u tgbotapi.Update) {
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

		bot.SendMsg(chatID, fmt.Sprintf("CPU: %.2f%%\nMemory: %.2f%%",
			cpuPercent[0],
			memPercent.UsedPercent,
		))
	})

	bot.AddCmd("set", "Set config value", config.CmdSet)

	bot.AddCmd("config", "Get all config values", func(b *mybot.Bot, u tgbotapi.Update) {
		b.SendMsg(u.Message.Chat.ID, config.All())
	})

	bot.AddCmd("plot", "Plot resource usage", func(b *mybot.Bot, u tgbotapi.Update) {
		plot(b, u.Message.Chat.ID)
	})

	bot.AddCmd("add", "Manualy add data point (for debug)", func(b *mybot.Bot, u tgbotapi.Update) {
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

	bot.AddCmd("plotBtn", "Send a message with a button to exec plot command", func(b *mybot.Bot, u tgbotapi.Update) {
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

	bot.AddCmd("hi", "Example command for waiting", func(b *mybot.Bot, u tgbotapi.Update) {
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

	bot.AddCmd("cancel", "Cancel waiting", func(b *mybot.Bot, u tgbotapi.Update) {
		b.Cancel(u.Message.Chat.ID)
		b.SendMsg(u.Message.Chat.ID, "Cancelled")
	})
}

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

func isSuddenlyIncrease(curr float64, history *history.History) bool {
	avg := history.Average(10 * time.Minute)

	// TODO: This formula is not stable, it need to be checked
	if avg < 1 {
		return curr > avg*10
	} else if avg < 10 {
		return curr > avg*3
	} else if avg < 30 {
		return curr > avg*2
	} else if avg < 60 {
		return curr > avg*1.5
	} else {
		return curr > avg*1.2
	}
}

func checkAndNotify(bot *mybot.Bot) {
	cpuPercent, err := cpu.Percent(time.Second, false)

	avgInterval := 10 * time.Minute

	if err == nil {
		if cpuPercent[0] > config.GetFloat64("cpu_threshold") {
			bot.Boradcast(fmt.Sprintf("High CPU usage detected: %.2f%%", cpuPercent[0]))
		}

		if isSuddenlyIncrease(cpuPercent[0], cpuUsageHistory) {
			avgCPU := cpuUsageHistory.Average(avgInterval)
			bot.Boradcast(fmt.Sprintf("Sudden increase in CPU usage detected: %.2f%% (Avg: %.2f%%)", cpuPercent[0], avgCPU))
		}

		cpuUsageHistory.Append(cpuPercent[0])
	}

	memStat, err := mem.VirtualMemory()
	if err == nil {
		if memStat.UsedPercent > config.GetFloat64("mem_threshold") {
			bot.Boradcast(fmt.Sprintf("High memory usage detected: %.2f%%", memStat.UsedPercent))
		}

		if isSuddenlyIncrease(memStat.UsedPercent, memUsageHistory) {
			avgMem := memUsageHistory.Average(avgInterval)
			bot.Boradcast(fmt.Sprintf("Sudden increase in memory usage detected: %.2f%% (Avg: %.2f%%)", memStat.UsedPercent, avgMem))
		}

		memUsageHistory.Append(memStat.UsedPercent)
	}
}
