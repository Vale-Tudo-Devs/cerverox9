package utils

import (
	"github.com/bwmarrin/discordgo"
)

func UserDisplayName(m *discordgo.Member) string {
	switch {
	case m.Nick != "":
		return m.Nick
	case m.User.GlobalName != "":
		return m.User.GlobalName
	default:
		return m.User.Username
	}
}
