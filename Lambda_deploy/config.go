package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

// DevModeを定義（テスト時は開発モードを有効にする）
const DevMode = true

var DiscordToken string
var TextChannelID string
var CollectionName string

// DevModeの設定を集約する関数
func SetupDevMode() {
	// 環境モードに応じた設定を行う
	var envFile string
	if DevMode {
		envFile = ".env.dev"  // 開発環境用
		CollectionName = "test_profiles" // 開発用コレクション
		fmt.Println("現在、開発モードで実行中です。")
	} else {
		envFile = ".env.prod" // 本番環境用
		CollectionName = "user_profiles" // 本番用コレクション
		fmt.Println("現在、本番モードで実行中です。")
	}

	// 環境変数をロード
	err := godotenv.Load(envFile)
	if err != nil {
		log.Fatalf("%sの読み込みに失敗しました: %v", envFile, err)
	}

	// Discordトークンの取得
	DiscordToken = os.Getenv("DISCORDTOKEN")
	if DiscordToken == "" {
		log.Fatal("環境変数DISCORDTOKENが設定されていません")
	}

	// 環境モードに応じたメッセージ送信先のチャンネルID指定
	TextChannelID = os.Getenv("DISCORDTEXTCHANNELID")
	if TextChannelID == "" {
		log.Fatal("環境変数DISCORDTEXTCHANNELIDが設定されていません")
	}
}