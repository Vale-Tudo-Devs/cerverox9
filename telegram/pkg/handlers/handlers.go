package handlers

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/vcaldo/cerverox9/telegram/pkg/stats"
)

func StatusHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	guildName, oncallUsersCount, oncallUsers, onlineUsersCount, onlineUsers, err := stats.GetVoiceCallStatus()
	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Error fetching voice call status",
		})
		return
	}

	oncallUsersList := strings.Split(oncallUsers, ",")
	oncallUsersListLinebreak := strings.Join(oncallUsersList, "\n")
	onlineUsersList := strings.Split(onlineUsers, ",")
	onlineUsersListLinebreak := strings.Join(onlineUsersList, "\n")
	discordInviteLink := os.Getenv("DISCORD_INVITE_LINK")

	message := fmt.Sprintf(
		"Live stats for Discord Server %s\n\n"+
			"We have %d users having fun in the call\n\n"+
			"%s\n\n"+
			"There are %d users who are one click away from having fun\n\n"+
			"%s\n\n"+
			"🥳 Join the party! 🥳\n%s",
		guildName,
		oncallUsersCount,
		oncallUsersListLinebreak,
		onlineUsersCount,
		onlineUsersListLinebreak,
		discordInviteLink,
	)

	var emojis = []string{
		"🧉", "🆙", "🫂", "🥃", "🆒", "🐞", "📱", "🎙", "🪩", "🎤", "🌡", "👽", "🦬", "🎢", "📞", "☎️", "💥", "🪙", "💃", "🕺", "💬", "🔥", "🎊", "👍🏿", "🥫", "🦾", "🧽", "🥰", "🧮", "🚑", "🧻", "🫰", "🤙", "🍙", "💪", "🙏", "🤲", "🫡", "🗣", "🦷", "💅",
	}
	var emptyEmojis = []string{
		"🫥", "⚰️", "🦠", "🙊", "😴", "😤", "🤬", "🥶", "🧟", "🕸", "☠️", "💤", "❄️", "😶", "🤚", "😓", "😫", "💩", "🤐", "🕊", "🗝", "🤨", "👹", "👺", "🫠", "😶‍🌫️", "😵", "🙉", "🦴", "🎟", "🏴", "⛈", "🤦‍♂️", "🦟", "🦝", "🖕", "💔", "🫵", "🤰", "🦍",
	}
	emojiMessage := emojis[rand.Intn(len(emojis))]
	if oncallUsersCount == 0 {
		emojiMessage = emptyEmojis[rand.Intn(len(emptyEmojis))]
	}
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   emojiMessage,
	})

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:             update.Message.Chat.ID,
		Text:               message,
		LinkPreviewOptions: &models.LinkPreviewOptions{IsDisabled: bot.True()},
	})
}

func VoiceEventHanlder(ctx context.Context, b *bot.Bot, event *VoiceEvent) {
	chatId, ok := os.LookupEnv("TELEGRAM_CHAT_ID")
	if !ok {
		panic("TELEGRAM_CHAT_IDenv var is required")
	}

	chatIdInt, err := strconv.ParseInt(chatId, 10, 64)
	if err != nil {
		panic("TELEGRAM_CHAT_ID must be a valid int64")
	}

	switch {
	// User joined the voice channel
	case event.EventType == "voice" && event.State:
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatIdInt,
			Text:   fmt.Sprintf("%s joined %s 🏃‍♂️", event.UserGlobalName, event.ChannelName),
		})
	// User left the voice channel
	case event.EventType == "voice" && !event.State:
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatIdInt,
			Text:   fmt.Sprintf("%s left %s 🏃‍♂️‍➡️", event.UserGlobalName, event.ChannelName),
		})
	case event.EventType == "webcam" && event.State:
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatIdInt,
			Text:   fmt.Sprintf("%s opened the webcam in %s 📸", event.UserGlobalName, event.ChannelName),
		})
	case event.EventType == "streaming" && event.State:
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatIdInt,
			Text:   fmt.Sprintf("%s started streaming in %s 📺 ", event.UserGlobalName, event.ChannelName),
		})
	}
}

func UserStatsHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	messageText := update.Message.Text
	words := strings.Fields(messageText)
	var targetUser string
	if len(words) == 1 {
		// Get the first word after /stats
		targetUser = words[1]
	} else {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Please provide a valid username",
		})
		return
	}

	userStats, err := stats.GetUserVoiceCallStatus(targetUser)
	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Error fetching user stats",
		})
		return
	}

	message := fmt.Sprintf(
		"📊 User stats\n\n"+
			"Total on call time this year for %s: %d\n",
		targetUser, userStats,
	)

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   message,
	})
}
