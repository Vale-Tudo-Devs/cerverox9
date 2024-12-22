package handlers

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/vcaldo/cerverox9/telegram/pkg/stats"
)

func StatusHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	oncallUsersCount, oncallUsers, onlineUsersCount, onlineUsers, err := stats.GetVoiceCallStatus()
	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Error fetching voice call status",
		})
		return
	}

	if len(oncallUsers) > 31 {
		oncallUsers = oncallUsers[31:]
	}
	oncallUsersList := strings.Split(oncallUsers, ",")
	oncallUsersListLinebreak := strings.Join(oncallUsersList, "\n")
	onlineUsersList := strings.Split(onlineUsers, ",")
	onlineUsersListLinebreak := strings.Join(onlineUsersList, "\n")

	discordInviteLink := os.Getenv("DISCORD_INVITE_LINK")

	message := fmt.Sprintf(
		"We have %d users having fun in the call.\n\n"+
			"%s\n\n"+
			"There are %d users who are one click away from having fun.\n\n"+
			"%s\n\n"+
			"🥳 Join the party! 🥳\n%s",
		oncallUsersCount,
		oncallUsersListLinebreak,
		onlineUsersCount,
		onlineUsersListLinebreak,
		discordInviteLink,
	)

	if oncallUsersCount == 0 {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "🤬",
		})
	} else {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "📞",
		})
	}

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
			Text:   fmt.Sprintf("%s joined the call", event.Username),
		})
	// User left the voice channel
	case event.EventType == "voice" && !event.State:
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatIdInt,
			Text:   fmt.Sprintf("%s left the call", event.Username),
		})
	case event.EventType == "webcam" && event.State:
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatIdInt,
			Text:   fmt.Sprintf("%s opened the webcam 📸", event.Username),
		})
	case event.EventType == "streaming" && event.State:
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatIdInt,
			Text:   fmt.Sprintf("%s started streaming 📺", event.Username),
		})
	}
}
