package main

import (
	"fmt"
	"log"
	mybot "monitor/bot"
	cfg "monitor/config"
	"monitor/history"
	"os"
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
	cpuUsageHistory = history.New(10) // History of CPU usage percentages
	memUsageHistory = history.New(10) // History of memory usage percentages
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

		bot.SendMsg(chatID, fmt.Sprintf("===Usage===\nCPU:%.2f%%\nMemory: %.2f%%\n\n===Config===\nCPU Threshold: %.2f%%\nMemory Threshold: %.2f%%\nIncrease Threshold: %.2f\nInterval: %d minutes",
			cpuPercent[0],
			memPercent.UsedPercent,
			config.GetFloat64("cpu_threshold"),
			config.GetFloat64("mem_threshold"),
			config.GetFloat64("increase_threshold"),
			config.GetInt("interval"),
		))
	})

	bot.AddCmd("set", "Set config value", func(b *mybot.Bot, u tgbotapi.Update) {
		if !config.Cmd(u) {
			b.SendMsg(u.Message.Chat.ID, "Invalid config")
		}
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
	if err == nil {
		if cpuPercent[0] > config.GetFloat64("cpu_threshold") {
			bot.Boradcast(fmt.Sprintf("High CPU usage detected: %.2f%%", cpuPercent[0]))
		}

		cpuUsageHistory.Append(cpuPercent[0])
		avgCPU := cpuUsageHistory.Average()
		if cpuPercent[0] > avgCPU*config.GetFloat64("increase_threshold") {
			bot.Boradcast(fmt.Sprintf("Sudden increase in CPU usage detected: %.2f%% (Avg: %.2f%%)", cpuPercent[0], avgCPU))
		}
	}

	memStat, err := mem.VirtualMemory()
	if err == nil {
		if memStat.UsedPercent > config.GetFloat64("mem_threshold") {
			bot.Boradcast(fmt.Sprintf("High memory usage detected: Total: %v MB, Used: %v MB, Usage: %.2f%%",
				memStat.Total/1024/1024, memStat.Used/1024/1024, memStat.UsedPercent))
		}

		memUsageHistory.Append(memStat.UsedPercent)
		avgMem := memUsageHistory.Average()
		if memStat.UsedPercent > avgMem*config.GetFloat64("increase_threshold") {
			bot.Boradcast(fmt.Sprintf("Sudden increase in memory usage detected: %.2f%% (Avg: %.2f%%)", memStat.UsedPercent, avgMem))
		}
	}
}
