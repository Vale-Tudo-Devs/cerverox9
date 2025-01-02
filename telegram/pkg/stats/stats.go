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
	duration, err := dm.GetUserVoiceTime(username)
	if err != nil {
		return 0, err
	}

	return duration, nil
}
