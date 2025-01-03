package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/vcaldo/cerverox9/telegram/pkg/handlers"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	opts := []bot.Option{
		bot.WithDefaultHandler(handler),
	}

	token, ok := os.LookupEnv("TELEGRAM_BOT_TOKEN")
	if !ok {
		log.Fatal("TELEGRAM_BOT_TOKEN env var is required")
	}
	b, err := bot.New(token, opts...)
	if err != nil {
		panic(err)
	}

	// Start the bot in a goroutine
	go func() {
		b.Start(ctx)
	}()

	// Start the voice event listener
	listener := handlers.NewVoiceEventListener()
	go func() {
		listener.Start(ctx)
	}()

	// Listen for events from the voice channel and send messages
	go func() {
		for event := range listener.NotifyChan {
			handlers.VoiceEventHanlder(ctx, b, &event)
		}
	}()

	// Wait for the context to be done
	select {}
}

func handler(ctx context.Context, b *bot.Bot, update *models.Update) {
	switch {
	case update.Message != nil && update.Message.Text == "/status":
		handlers.StatusHandler(ctx, b, update)
	case update.Message != nil && strings.HasPrefix(update.Message.Text, "/voicestats"):
		handlers.UserStatsHandler(ctx, b, update)
	}
}
