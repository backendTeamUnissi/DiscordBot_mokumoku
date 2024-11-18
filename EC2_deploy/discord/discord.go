// discord.go
package discord

import (
    "time"

    "DiscordBot_mokumoku/EC2_deploy/config"
    firestorePkg "DiscordBot_mokumoku/EC2_deploy/firestore"
    "github.com/bwmarrin/discordgo"
    firestore "cloud.google.com/go/firestore"
)

func InitDiscord(token string) (*discordgo.Session, error) {
    dg, err := discordgo.New("Bot " + token)
    if err != nil {
        return nil, err
    }

    dg.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildVoiceStates
    return dg, nil
}

func RegisterHandlers(s *discordgo.Session, fsClient *firestore.Client, cfg *config.Config) {
    fsService := firestorePkg.NewFirestoreService(fsClient)
    handler := &DiscordHandler{
        FirestoreService: fsService,
        TextChannelID:    cfg.DiscordTextChannelID,
        VoiceChannelID:   cfg.DiscordVoiceChannelID,
        UserJoinTimes:    make(map[string]time.Time),
    }
    s.AddHandler(handler.VoiceStateUpdate)
}
