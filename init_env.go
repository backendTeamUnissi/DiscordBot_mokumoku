package main
import (
	"log"
	"os"

	"github.com/joho/godotenv"
)
func init_env(){
	err := godotenv.Load()
	if err != nil {
		log.Fatalf(".envファイルの読み込みに失敗しました: %v", err)
	}

	// ここで初期化
	token = os.Getenv("DISCORDTOKEN")
	if token == "" {
		log.Fatal("Discordトークンが設定されていません。環境変数DISCORDTOKENを設定してください。")
	}

	textChannelID = os.Getenv("DISCORDTEXTCHANNELID")
	if textChannelID == "" {
		log.Fatal("DiscordチャンネルIDが設定されていません。環境変数DISCORDTEXTCHANNELIDを設定してください。")
	}

	voiceChannelID = os.Getenv("DISCORDVOICECHANNELID")
	if voiceChannelID == "" {
		log.Fatal("DiscordボイスチャンネルIDが設定されていません。環境変数DISCORDVOICECHANNELIDを設定してください。")
	}
}