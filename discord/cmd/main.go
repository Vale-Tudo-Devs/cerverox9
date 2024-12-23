package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/vcaldo/cerverox9/discord/pkg/handlers"
	"github.com/vcaldo/cerverox9/discord/pkg/models"
)

func main() {
	ctx := context.Background()
	token, ok := os.LookupEnv("DISCORD_BOT_TOKEN")
	if !ok {
		log.Fatal("DISCORD_BOT_TOKEN env var is required")
	}

	dg, err := discordgo.New(fmt.Sprintf("Bot %s", token))
	if err != nil {
		log.Println("error creating Discord session,", err)
		return
	}

	// Register necessary Intents for the bot
	dg.Identify.Intents = discordgo.IntentGuilds |
		discordgo.IntentsGuildPresences |
		discordgo.IntentGuildMembers |
		discordgo.IntentGuildVoiceStates

	dg.AddHandler(handlers.VoiceStateUpdate)

	err = dg.Open()
	if err != nil {
		log.Println("error opening connection,", err)
		return
	}

	log.Println("Discord Bot is now running.")

	// Launch a goroutine to update user presence when the bot starts
	dm := models.NewAuthenticatedDiscordMetricsClient()
	go dm.LogUsersPresence(dg)

	// Update user presence every 60 seconds
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				dm.LogUsersPresence(dg)
			}
		}
	}()

	select {}
}
