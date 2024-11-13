package discord

import (
	"github.com/bwmarrin/discordgo"
)

func InitDiscord(token string) (*discordgo.Session, error) {
	return discordgo.New("Bot " + token)
}
