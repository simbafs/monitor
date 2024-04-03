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

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

var telegramBotToken = os.Getenv("TG_BOT_TOKEN")

var config = cfg.New().Float64("cpu_threshold", "CPU threshold", 75.0).
	Float64("mem_threshold", "Memory threshold", 85.0).
	Float64("increase_threshold", "Increase threshold", 1.25). // TODO: thie should be a function of precentage of resource
	Int("interval", "Interval", 1)

var (
	cpuUsageHistory = history.New(10*time.Minute, "CPU")    // History of CPU usage percentages
	memUsageHistory = history.New(10*time.Minute, "Memory") // History of memory usage percentages
)

func main() {
	bot, err := mybot.New(tgbotapi.NewBotAPI(telegramBotToken))
	if err != nil {
		log.Fatal(err)
	}

	registerCmds(bot)

	// bot.Debug = true

	bot.Boradcast("Bot started")
	fmt.Println("Bot started")

	go handleCommands(bot)

	for {
		checkAndNotify(bot)
		time.Sleep(time.Duration(config.GetInt("interval")) * time.Minute) // Check every 5 minutes
	}
}

func registerCmds(bot *mybot.Bot) {
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
		img, err := history.Plot(cpuUsageHistory, memUsageHistory)
		if err != nil {
			b.SendMsg(u.Message.Chat.ID, "Error plotting")
		}

		var imgBuf bytes.Buffer
		if _, err := img.WriteTo(&imgBuf); err != nil {
			b.SendMsg(u.Message.Chat.ID, "Error plotting")
		}

		file := tgbotapi.FileBytes{Name: "usage.png", Bytes: imgBuf.Bytes()}

		photo := tgbotapi.NewPhotoUpload(u.Message.Chat.ID, file)

		b.Send(photo)
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
}

func handleCommands(bot *mybot.Bot) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.Bot.GetUpdatesChan(u)
	if err != nil {
		panic(err)
	}

	for update := range updates {
		bot.HandleCmds(update)
	}
}

func checkAndNotify(bot *mybot.Bot) {
	cpuPercent, err := cpu.Percent(time.Second, false)

	interval := 10 * time.Minute

	if err == nil {
		if cpuPercent[0] > config.GetFloat64("cpu_threshold") {
			bot.Boradcast(fmt.Sprintf("High CPU usage detected: %.2f%%", cpuPercent[0]))
		}

		cpuUsageHistory.Append(cpuPercent[0])
		avgCPU := cpuUsageHistory.Average(interval)
		if cpuPercent[0] > avgCPU*config.GetFloat64("increase_threshold") {
			bot.Boradcast(fmt.Sprintf("Sudden increase in CPU usage detected: %.2f%% (Avg: %.2f%%)", cpuPercent[0], avgCPU))
		}
	}

	memStat, err := mem.VirtualMemory()
	if err == nil {
		if memStat.UsedPercent > config.GetFloat64("mem_threshold") {
			bot.Boradcast(fmt.Sprintf("High memory usage detected: %.2f%%", memStat.UsedPercent))
		}

		memUsageHistory.Append(memStat.UsedPercent)
		avgMem := memUsageHistory.Average(interval)
		if memStat.UsedPercent > avgMem*config.GetFloat64("increase_threshold") {
			bot.Boradcast(fmt.Sprintf("Sudden increase in memory usage detected: %.2f%% (Avg: %.2f%%)", memStat.UsedPercent, avgMem))
		}
	}
}
