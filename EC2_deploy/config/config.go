// config.go
package config

import (
    "fmt"
    "log"
    "os"

    "github.com/joho/godotenv"
)

type Config struct {
    DiscordToken          string
    DiscordTextChannelID  string
    DiscordVoiceChannelID string
    FirestoreCredentials  string
}

func LoadConfig() (*Config, error) {
    // .env ファイルの読み込み
    err := godotenv.Load()
    if err != nil {
        log.Println("No .env file found, reading environment variables")
    }

    cfg := &Config{
        DiscordToken:          os.Getenv("DISCORDTOKEN"),
        DiscordTextChannelID:  os.Getenv("DISCORDTEXTCHANNELID"),
        DiscordVoiceChannelID: os.Getenv("DISCORDVOICECHANNELID"),
        FirestoreCredentials:  os.Getenv("FIRESTORE_CREDENTIALS_FILE"),
    }

    // 環境変数のバリデーション
    if cfg.DiscordToken == "" || cfg.DiscordTextChannelID == "" || cfg.DiscordVoiceChannelID == "" || cfg.FirestoreCredentials == "" {
        return nil, fmt.Errorf("one or more required environment variables are missing")
    }

    return cfg, nil
}
