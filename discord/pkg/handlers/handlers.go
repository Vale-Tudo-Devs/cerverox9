package handlers

import (
	"log"
	"os"
	"slices"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/vcaldo/cerverox9/discord/pkg/models"
)

func VoiceStateUpdate(s *discordgo.Session, vsu *discordgo.VoiceStateUpdate) {
	dm := models.NewAuthenticatedDiscordMetricsClient()
	ignoredUsers := strings.Split(os.Getenv("DISCORD_IGNORED_USERNAMES"), ",")
	if slices.Contains(ignoredUsers, vsu.Member.User.Username) {
		return
	}
	switch {
	// User joined a voice channel
	case vsu.BeforeUpdate == nil && vsu.ChannelID != "":
		user, err := s.User(vsu.UserID)
		if err != nil {
			log.Println("error fetching user:", err)
			return
		}
		log.Printf("User %s has joined voice channel %s", user.Username, vsu.ChannelID)

		err = dm.LogVoiceEvent(s, vsu, vsu.ChannelID, models.VoiceEvent, true)
		if err != nil {
			log.Println("error logging voice event:", err)
			return
		}

		// Update the number of users in voice channels when a user joins
		err = dm.LogUsersPresence(s)
		if err != nil {
			log.Println("error register users in voice channels:", err)
		}
	// User left a voice channel
	case vsu.ChannelID == "":
		user, err := s.User(vsu.UserID)
		if err != nil {
			log.Println("error fetching user:", err)
			return
		}
		log.Printf("User %s has left voice channel %s", user.Username, vsu.BeforeUpdate.ChannelID)

		err = dm.LogVoiceEvent(s, vsu, vsu.BeforeUpdate.ChannelID, models.VoiceEvent, false)
		if err != nil {
			log.Println("error logging voice event:", err)
			return
		}

		// Update the number of users in voice channels when a user joins
		err = dm.LogUsersPresence(s)
		if err != nil {
			log.Println("error register users in voice channels:", err)
		}
	// User switched voice channels
	case vsu.ChannelID != "" && vsu.BeforeUpdate.ChannelID != vsu.ChannelID:
		user, err := s.User(vsu.UserID)
		if err != nil {
			log.Println("error fetching user:", err)
			return
		}
		log.Printf("User %s has switched from voice channel %s to %s", user.Username, vsu.BeforeUpdate.ChannelID, vsu.ChannelID)

		// When user swtiches channels, they leave the previous one and join the new one
		err = dm.LogVoiceEvent(s, vsu, vsu.BeforeUpdate.ChannelID, models.VoiceEvent, false)
		if err != nil {
			log.Println("error logging voice event:", err)
			return
		}

		err = dm.LogVoiceEvent(s, vsu, vsu.ChannelID, models.VoiceEvent, true)
		if err != nil {
			log.Println("error register users in voice channels:", err)
		}

		// Update the number of users swtich channels
		err = dm.LogUsersPresence(s)
		if err != nil {
			log.Println("error register users in voice channels:", err)
		}
	// User started streaming
	case !vsu.BeforeUpdate.SelfStream && vsu.SelfStream:
		user, err := s.User(vsu.UserID)
		if err != nil {
			log.Println("error fetching user:", err)
			return
		}
		log.Printf("User %s has started streaming in voice channel %s", user.Username, vsu.ChannelID)

		err = dm.LogVoiceEvent(s, vsu, vsu.ChannelID, models.StreamEvent, true)
		if err != nil {
			log.Println("error logging voice event:", err)
			return
		}
	// User stopped streaming
	case vsu.BeforeUpdate.SelfStream && !vsu.SelfStream:
		user, err := s.User(vsu.UserID)
		if err != nil {
			log.Println("error fetching user:", err)
			return
		}
		log.Printf("User %s has stopped streaming in voice channel %s", user.Username, vsu.ChannelID)

		err = dm.LogVoiceEvent(s, vsu, vsu.ChannelID, models.StreamEvent, false)
		if err != nil {
			log.Println("error logging voice event:", err)
			return
		}
	// User turned their webcam on
	case !vsu.BeforeUpdate.SelfVideo && vsu.SelfVideo:
		user, err := s.User(vsu.UserID)
		if err != nil {
			log.Println("error fetching user:", err)
			return
		}
		log.Printf("User %s has turned on their webcam in voice channel %s", user.Username, vsu.ChannelID)

		err = dm.LogVoiceEvent(s, vsu, vsu.ChannelID, models.WebcamEvent, true)
		if err != nil {
			log.Println("error logging voice event:", err)
			return
		}
	// User turned their webcam off
	case vsu.BeforeUpdate.SelfVideo && !vsu.SelfVideo:
		user, err := s.User(vsu.UserID)
		if err != nil {
			log.Println("error fetching user:", err)
			return
		}
		log.Printf("User %s has turned off their webcam in voice channel %s", user.Username, vsu.ChannelID)

		err = dm.LogVoiceEvent(s, vsu, vsu.ChannelID, models.WebcamEvent, false)
		if err != nil {
			log.Println("error logging voice event:", err)
			return
		}
	// User muted themselves
	case !vsu.BeforeUpdate.SelfMute && vsu.SelfMute:
		user, err := s.User(vsu.UserID)
		if err != nil {
			log.Println("error fetching user:", err)
			return
		}
		log.Printf("User %s has muted themselves in voice channel %s", user.Username, vsu.ChannelID)

		err = dm.LogVoiceEvent(s, vsu, vsu.ChannelID, models.MuteEvent, true)
		if err != nil {
			log.Println("error logging voice event:", err)
			return
		}
	// User unmuted themselves
	case vsu.BeforeUpdate.SelfMute && !vsu.SelfMute:
		user, err := s.User(vsu.UserID)
		if err != nil {
			log.Println("error fetching user:", err)
			return
		}
		log.Printf("User %s has unmuted themselves in voice channel %s", user.Username, vsu.ChannelID)

		err = dm.LogVoiceEvent(s, vsu, vsu.ChannelID, models.MuteEvent, false)
		if err != nil {
			log.Println("error logging voice event:", err)
			return
		}
	// User deafened themselves
	case !vsu.BeforeUpdate.SelfDeaf && vsu.SelfDeaf:
		user, err := s.User(vsu.UserID)
		if err != nil {
			log.Println("error fetching user:", err)
			return
		}
		log.Printf("User %s has deafened themselves in voice channel %s", user.Username, vsu.ChannelID)

		err = dm.LogVoiceEvent(s, vsu, vsu.ChannelID, models.DeafenEvent, true)
		if err != nil {
			log.Println("error logging voice event:", err)
			return
		}
	// User undeafened themselves
	case vsu.BeforeUpdate.SelfDeaf && !vsu.SelfDeaf:
		user, err := s.User(vsu.UserID)
		if err != nil {
			log.Println("error fetching user:", err)
			return
		}
		log.Printf("User %s has undeafened themselves in voice channel %s", user.Username, vsu.ChannelID)

		err = dm.LogVoiceEvent(s, vsu, vsu.ChannelID, models.DeafenEvent, false)
		if err != nil {
			log.Println("error logging voice event:", err)
			return
		}
	}
}
