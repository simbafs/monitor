package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

type Config struct {
	CPUThreshold float64 `json:"cpu_threshold"` // persentage
	MemThreshold float64 `json:"mem_threshold"` // persentage

	HistoryLength      int     `json:"history_length"`
	InscreaseThreshold float64 `json:"increase_threshold"`

	Interval int `json:"interval"` // in minutes
}

var telegramBotToken = os.Getenv("TG_BOT_TOKEN")

var chatIDs = map[int64]bool{
	901756183: true,
}

var (
	cpuUsageHistory []float64 // History of CPU usage percentages
	memUsageHistory []float64 // History of memory usage percentages
)

var config = Config{
	CPUThreshold:       75.0,
	MemThreshold:       85.0,
	HistoryLength:      10,
	InscreaseThreshold: 1.25,
	Interval:           1,
}

func main() {
	bot, err := tgbotapi.NewBotAPI(telegramBotToken)
	if err != nil {
		log.Fatal(err)
	}

	// bot.Debug = true

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	sendMessage(bot, "Bot started")

	go handleCommands(bot, updates)

	for {
		checkAndNotify(bot)
		time.Sleep(time.Duration(config.Interval) * time.Minute) // Check every 5 minutes
	}
}

func handleCommands(bot *tgbotapi.BotAPI, updates tgbotapi.UpdatesChannel) {
	for update := range updates {
		if update.Message == nil || !update.Message.IsCommand() {
			continue
		}
		switch update.Message.Command() {
		case "subscribe":
			chatID := update.Message.Chat.ID
			if chatIDs[chatID] {
				sendMessage(bot, "Already subscribed")
			} else {
				chatIDs[chatID] = true
				sendMessage(bot, "Subscribed to notifications")
			}
		case "unsubscribe":
			chatID := update.Message.Chat.ID
			if chatIDs[chatID] {
				sendMessage(bot, "Unsubscribed from notifications")
				delete(chatIDs, chatID)
			} else {
				sendMessage(bot, "Not subscribed")
			}
		case "setcpu":
			if newThreshold, err := strconv.ParseFloat(update.Message.CommandArguments(), 64); err == nil && newThreshold >= 0 {
				config.CPUThreshold = newThreshold
				sendMessage(bot, fmt.Sprintf("CPU threshold set to: %.2f%%", config.CPUThreshold))
			} else {
				sendMessage(bot, "Error setting CPU threshold. Please use a valid number.")
			}
		case "setmem":
			if newThreshold, err := strconv.ParseFloat(update.Message.CommandArguments(), 64); err == nil && newThreshold >= 0 {
				config.MemThreshold = newThreshold
				sendMessage(bot, fmt.Sprintf("Memory threshold set to: %.2f%%", config.MemThreshold))
			} else {
				sendMessage(bot, "Error setting memory threshold. Please use a valid number.")
			}
		case "setinterval":
			if newInterval, err := strconv.Atoi(update.Message.CommandArguments()); err == nil && newInterval > 0 {
				config.Interval = newInterval
				sendMessage(bot, fmt.Sprintf("Interval set to: %d minutes", config.Interval))
			} else {
				sendMessage(bot, "Error setting interval. Please use a valid number.")
			}
		case "setIncreaseThreshold":
			if newThreshold, err := strconv.ParseFloat(update.Message.CommandArguments(), 64); err == nil && newThreshold >= 0 {
				config.InscreaseThreshold = newThreshold
				sendMessage(bot, fmt.Sprintf("Increase threshold set to: %.2f%%", config.InscreaseThreshold))
			} else {
				sendMessage(bot, "Error setting increase threshold. Please use a valid number.")
			}
		case "status":
			cpuPercent, err := cpu.Percent(time.Second, false)
			if err != nil {
				sendMessage(bot, "Error getting CPU usage")
				continue
			}
			memPercent, err := mem.VirtualMemory()
			if err != nil {
				sendMessage(bot, "Error getting memory usage")
				continue
			}

			sendMessage(bot, fmt.Sprintf("===Usage===\nCPU:%.2f%%\nMemory: %.2f%%\n\n===Config===\nCPU Threshold: %.2f%%\nMemory Threshold: %.2f%%\nInterval: %d minutes",
				cpuPercent[0],
				memPercent.UsedPercent,
				config.CPUThreshold,
				config.MemThreshold,
				config.Interval,
			))
		default:
			// send commands list to user
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Available commands:\n"+
				"/subscribe - Subscribe to notifications\n"+
				"/unsubscribe - Unsubscribe from notifications\n"+
				"/setcpu - Set CPU threshold\n"+
				"/setmem - Set memory threshold\n"+
				"/setinterval - Set interval\n"+
				"/setIncreaseThreshold - Set increase threshold\n"+
				"/status - Get current status")
			bot.Send(msg)
				
		}

	}
}

func checkAndNotify(bot *tgbotapi.BotAPI) {
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err == nil {
		if cpuPercent[0] > config.CPUThreshold {
			sendMessage(bot, fmt.Sprintf("High CPU usage detected: %.2f%%", cpuPercent[0]))
		}

		updateHistory(&cpuUsageHistory, cpuPercent[0])
		avgCPU := average(cpuUsageHistory)
		if cpuPercent[0] > avgCPU*config.InscreaseThreshold {
			sendMessage(bot, fmt.Sprintf("Sudden increase in CPU usage detected: %.2f%% (Avg: %.2f%%)", cpuPercent[0], avgCPU))
		}
	}

	memStat, err := mem.VirtualMemory()
	if err == nil {
		if memStat.UsedPercent > config.MemThreshold {
			sendMessage(bot, fmt.Sprintf("High memory usage detected: Total: %v MB, Used: %v MB, Usage: %.2f%%",
				memStat.Total/1024/1024, memStat.Used/1024/1024, memStat.UsedPercent))
		}

		updateHistory(&memUsageHistory, memStat.UsedPercent)
		avgMem := average(memUsageHistory)
		if memStat.UsedPercent > avgMem*config.InscreaseThreshold {
			sendMessage(bot, fmt.Sprintf("Sudden increase in memory usage detected: %.2f%% (Avg: %.2f%%)", memStat.UsedPercent, avgMem))
		}
	}
}

func updateHistory(history *[]float64, newValue float64) {
	if len(*history) >= config.HistoryLength {
		*history = (*history)[1:]
	}
	*history = append(*history, newValue)
}

func average(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func sendMessage(bot *tgbotapi.BotAPI, message string) {
	for chatID := range chatIDs {
		msg := tgbotapi.NewMessage(chatID, message)
		_, err := bot.Send(msg)
		if err != nil {
			log.Printf("Error sending message: %s\n", err)
		}
	}
}
