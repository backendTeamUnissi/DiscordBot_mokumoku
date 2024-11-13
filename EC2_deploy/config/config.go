package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DiscordToken         string
	DiscordTextChannelID string
	DiscordVoiceChannelID string
	FirestoreCredentials string
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, reading environment variables")
	}

	return &Config{
		DiscordToken:         os.Getenv("DISCORDTOKEN"),
		DiscordTextChannelID: os.Getenv("DISCORDTEXTCHANNELID"),
		DiscordVoiceChannelID: os.Getenv("DISCORDVOICECHANNELID"),
		FirestoreCredentials: os.Getenv("FIRESTORE_CREDENTIALS_FILE"), // .envに追加
	}, nil
}
