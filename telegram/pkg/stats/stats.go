package stats

import (
	"log"
	"os"
	"time"

	"github.com/vcaldo/cerverox9/discord/pkg/models"
)

func GetVoiceCallStatus() (guildName string, oncallUsersCount int64, oncallUsers string, onlineUsersCount int64, onlineUsers string, error error) {
	dm := models.NewAuthenticatedDiscordMetricsClient()
	guildID, ok := os.LookupEnv("DISCORD_GUILD_ID")
	if !ok {
		log.Fatal("DISCORD_GUILD_ID env var is required")
	}

	guildName, oncallUsersCount, oncallUsers, err := dm.GetOncallUsers(guildID)
	if err != nil {
		return "", 0, "", 0, "", err
	}

	_, onlineUsersCount, onlineUsers, err = dm.GetOnlineUsers(guildID)
	if err != nil {
		return "", 0, "", 0, "", err
	}

	return guildName, oncallUsersCount, oncallUsers, onlineUsersCount, onlineUsers, nil
}

func GetUserVoiceCallStatus(username string) (time.Duration, error) {
	dm := models.NewAuthenticatedDiscordMetricsClient()

	guildID, ok := os.LookupEnv("DISCORD_GUILD_ID")
	if !ok {
		log.Fatal("DISCORD_GUILD_ID env var is required")
	}

	ignoredVoiceChannel, ok := os.LookupEnv("DISCORD_IGNORED_VOICE_TIME_COUNT_CHANNEL")
	if !ok {
		log.Fatal("DISCORD_IGNORED_VOICE_TIME_COUNT_CHANNEL env var is required")
	}

	duration, err := dm.GetUserVoiceTime(username, guildID, ignoredVoiceChannel)
	if err != nil {
		return 0, err
	}

	return duration, nil
}

func GetVoiceCallRank() (guildName, totalDuration, rank string, err error) {
	dm := models.NewAuthenticatedDiscordMetricsClient()
	guildID, ok := os.LookupEnv("DISCORD_GUILD_ID")
	if !ok {
		log.Fatal("DISCORD_GUILD_ID env var is required")
	}

	guildName, totalOncallDuration, voiceRank, err := dm.GetVoiceRank(guildID)
	if err != nil {
		return "", "", "", err
	}

	return guildName, totalOncallDuration, voiceRank, nil
}
