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
	defer dm.Close()

	go dm.LogUsersPresence(dg)

	// Update user presence every 30 seconds
	ticker := time.NewTicker(30 * time.Second)
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

	// Launch a goroutine to update user rank when the bot starts
	dm.UpdateVoiceRank(dg)

	// Update user rank every 5 minutes
	tickerRank := time.NewTicker(300 * time.Second)
	defer tickerRank.Stop()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-tickerRank.C:
				dm.UpdateVoiceRank(dg)
			}
		}
	}()

	guildId, ok := os.LookupEnv("DISCORD_GUILD_ID")
	if !ok {
		log.Fatal("DISCORD_GUILD_ID env var is required")
	}

	guildName, duration, voiceRank, err := dm.GetVoiceRank(guildId)
	if err != nil {
		log.Println("error getting voice rank", err)
		return
	}

	log.Printf("Guild: %s, Duration: %s, Voice Rank: %s\n", guildName, duration, voiceRank)

	select {}
}
