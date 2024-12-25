package models

import (
	"context"
	"fmt"
	"log"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

const (
	VoiceEventsMeasurement = "voice_events"
	OncallUsersMeasurement = "oncall_users"
	OnlineUsersMeasurement = "online_users"
	UserIdKey              = "user_id"
	UsernameKey            = "username"
	UserDisplayNameKey     = "user_display_name"
	GuildIdKey             = "guild_id"
	ChannelIdKey           = "channel_id"
	ChannelNameKey         = "channel_name"
	EventTypeKey           = "event_type"
	StateKey               = "state"
	VoiceEvent             = "voice"
	MuteEvent              = "mute"
	DeafenEvent            = "deafen"
	WebcamEvent            = "webcam"
	StreamEvent            = "streaming"
)

type DiscordMetrics struct {
	Client influxdb2.Client
	Org    string
	Bucket string
	Url    string
}

func NewAuthenticatedDiscordMetricsClient() *DiscordMetrics {
	influxUrl, ok := os.LookupEnv("INFLUX_URL")
	if !ok {
		log.Fatal("INFLUX_URL env var is required")
	}
	influxToken, ok := os.LookupEnv("INFLUX_TOKEN")
	if !ok {
		log.Fatal("INFLUX_TOKEN env var is required")
	}
	influxOrg, ok := os.LookupEnv("INFLUX_ORG")
	if !ok {
		log.Fatal("INFLUX_ORG env var is required")
	}
	influxBucket, ok := os.LookupEnv("INFLUX_BUCKET")
	if !ok {
		log.Fatal("INFLUX_BUCKET env var is required")
	}

	return newDiscordMetricsClient(
		influxUrl,
		influxToken,
		influxOrg,
		influxBucket,
	)
}

func newDiscordMetricsClient(url, token, org, bucket string) *DiscordMetrics {
	if !strings.HasPrefix(url, "http") {
		url = fmt.Sprintf("http://%s", url)
	}
	client := influxdb2.NewClient(url, token)
	return &DiscordMetrics{
		Client: client,
		Org:    org,
		Bucket: bucket,
		Url:    url,
	}
}

func (dm *DiscordMetrics) LogVoiceEvent(s *discordgo.Session, vsu *discordgo.VoiceStateUpdate, channelID, voiceEvent string, state bool) error {
	NewAuthenticatedDiscordMetricsClient()

	// Ignore users in the ignore list
	ignoredUsers := strings.Split(os.Getenv("DISCORD_IGNORED_USERNAMES"), ",")
	if slices.Contains(ignoredUsers, vsu.Member.User.Username) {
		return nil
	}

	user, err := s.User(vsu.UserID)
	if err != nil {
		return fmt.Errorf("error fetching user: %v", err)
	}

	channel, err := s.Channel(channelID)
	if err != nil {
		log.Println("error fetching channel:", err)
		return fmt.Errorf("error fetching channel: %v", err)
	}

	dm.logVoiceEvent(vsu.UserID, user.Username, userDisplayName(vsu.Member), vsu.GuildID, channelID, channel.Name, voiceEvent, state)
	return nil
}

func (dm *DiscordMetrics) logVoiceEvent(userID, username, userDisplayName, guildID, channelID, channelName, eventType string, state bool) error {
	writeAPI := dm.Client.WriteAPIBlocking(dm.Org, dm.Bucket)

	p := influxdb2.NewPoint(VoiceEventsMeasurement,
		map[string]string{
			UserIdKey:          userID,
			UsernameKey:        username,
			UserDisplayNameKey: userDisplayName,
			GuildIdKey:         guildID,
			ChannelIdKey:       channelID,
			ChannelNameKey:     channelName,
			EventTypeKey:       eventType,
		},
		map[string]interface{}{
			StateKey: state,
		},
		time.Now())
	log.Printf("Writing point: %s, %s, %s, %t in %s measurement", username, userDisplayName, eventType, state, VoiceEventsMeasurement)

	return writeAPI.WritePoint(context.Background(), p)
}

func (dm *DiscordMetrics) GetOncallUsers(guildID string) (int64, string, error) {
	// query oncall users
	query := fmt.Sprintf(`from(bucket:"%s")
		|> range(start: -10m)
		|> filter(fn: (r) => r._measurement == "%s" and r.guild_id == "%s")
		|> group(columns: ["guild_id"])
		|> sort(columns: ["_time"], desc: true)
		|> limit(n: 1)
		|> last()`,
		dm.Bucket, OncallUsersMeasurement, guildID)

	queryAPI := dm.Client.QueryAPI(dm.Org)
	result, err := queryAPI.Query(context.Background(), query)
	if err != nil {
		return 0, "", fmt.Errorf("error querying for oncall users: %v", err)
	}
	defer result.Close()

	for result.Next() {
		record := result.Record()
		oncallUsersCount := record.Value().(int64)
		var oncallUsers string
		if oncallUsersCount == 0 {
			oncallUsers = "Empty Discord. Crowded streets." // This can't have a comma
			return oncallUsersCount, oncallUsers, nil
		} else {
			oncallUsers = record.Values()["user_list"].(string)
			return oncallUsersCount, oncallUsers, nil
		}
	}
	return 0, "", fmt.Errorf("no online users found for guild %s", guildID)
}

func (dm *DiscordMetrics) GetOnlineUsers(guildID string) (int64, string, error) {
	// query online users
	query2 := fmt.Sprintf(`from(bucket:"%s")
		|> range(start: -10m)
		|> filter(fn: (r) => r._measurement == "%s" and r.guild_id == "%s")
		|> group(columns: ["guild_id"])
		|> sort(columns: ["_time"], desc: true)
		|> limit(n: 1)
		|> last()`,
		dm.Bucket, OnlineUsersMeasurement, guildID)

	queryAPI := dm.Client.QueryAPI(dm.Org)
	result, err := queryAPI.Query(context.Background(), query2)
	if err != nil {
		return 0, "", fmt.Errorf("error querying for online users: %v", err)
	}
	defer result.Close()

	for result.Next() {
		record := result.Record()
		onlineUsersCount := record.Value().(int64)
		var onlineUsers string
		if onlineUsersCount == 0 {
			onlineUsers = "" // This can't have a comma
			return onlineUsersCount, onlineUsers, nil
		} else {
			onlineUsers = record.Values()["user_list"].(string)
			return onlineUsersCount, onlineUsers, nil
		}
	}
	return 0, "", fmt.Errorf("no online users found for guild %s", guildID)
}

func (dm *DiscordMetrics) logUsersCount(measurementName, guildID string, userCount int, userList []string) error {
	writeAPI := dm.Client.WriteAPIBlocking(dm.Org, dm.Bucket)

	p := influxdb2.NewPoint(measurementName,
		map[string]string{
			"guild_id":  guildID,
			"user_list": strings.Join(userList, ","),
		},
		map[string]interface{}{
			"user_count": userCount,
		},
		time.Now())
	log.Printf("Writing point: %s, %d in %s measurement", guildID, userCount, measurementName)

	return writeAPI.WritePoint(context.Background(), p)
}

func (dm *DiscordMetrics) LogUsersPresence(s *discordgo.Session) error {
	guilds, err := s.UserGuilds(200, "", "", true)
	if err != nil {
		return fmt.Errorf("error fetching guilds: %v", err)
	}
	for _, guild := range guilds {
		// Register oncall users
		guildID := guild.ID
		members, err := s.GuildMembers(guildID, "", 1000)
		if err != nil {
			log.Printf("error fetching members for guild %s: %v", guildID, err)
			return fmt.Errorf("error fetching members for guild %s: %v", guildID, err)
		}
		oncallUsersCount := 0
		oncallUsers := []string{}
		for _, member := range members {
			if member.User.Bot {
				continue
			}
			vs, _ := s.State.VoiceState(guildID, member.User.ID) // it errors out if the user is not in a voice channel, ignore it
			if vs != nil && vs.ChannelID != "" {
				// Checj if the user is on an ignored channel
				currentVoiceChannel, err := s.Channel(vs.ChannelID)
				if err != nil {
					log.Printf("error fetching channel for user %s: %v", member.User.ID, err)
					continue
				}
				ignoredChannels := strings.Split(os.Getenv("DISCORD_IGNORED_CHANNELS"), ",")
				if slices.Contains(ignoredChannels, currentVoiceChannel.Name) {
					log.Printf("Ignoring user %s in ignored channel %s", member.User.ID, currentVoiceChannel.Name)
					continue
				}
				oncallUsersCount++
				oncallUsers = append(oncallUsers, userDisplayName(member))
			}
		}

		err = dm.logUsersCount(OncallUsersMeasurement, guildID, oncallUsersCount, oncallUsers)
		if err != nil {
			return fmt.Errorf("error logging online users: %v", err)
		}
		log.Printf("Logged %d on call users for guild %s - %s", oncallUsersCount, guildID, guild.Name)

		// Register online users
		onlineUsersCount := 0
		onlineUsers := []string{}
		for _, member := range members {
			if member.User.Bot {
				continue
			}
			presence, _ := s.State.Presence(guildID, member.User.ID) // it errors out if the user is not in a voice channel, ignore it
			if presence != nil && presence.Status != discordgo.StatusOffline {
				if !slices.Contains(oncallUsers, userDisplayName(member)) {
					onlineUsersCount++
					onlineUsers = append(onlineUsers, userDisplayName(member))
				}
			}
		}

		err = dm.logUsersCount(OnlineUsersMeasurement, guildID, onlineUsersCount, onlineUsers)
		if err != nil {
			return fmt.Errorf("error logging online users: %v", err)
		}
		log.Printf("Logged %d online users for guild %s - %s", onlineUsersCount, guildID, guild.Name)
	}
	return nil
}
