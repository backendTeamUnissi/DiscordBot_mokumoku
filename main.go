package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

// グローバル変数の宣言！（初期化はmain関数内で行う）
var token string
var textChannelID string
var voiceChannelID string

// userJoinTimes：ユーザーIDをキーに参加時刻を記録するマップ
var userJoinTimes = make(map[string]time.Time)

func main() {
	// .envファイルから環境変数を読み込み
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

	// DiscordAPIに接続するためのセッションを作成
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("Error creating Discord session: %v", err)
	}

	// Botを起動し、Discordサーバーに接続
	err = dg.Open()
	if err != nil {
		log.Fatalf("Error opening connection: %v", err)
	}

	// Botがシャットダウンされたときにセッションを閉じる
	defer dg.Close()
	fmt.Println("Bot is now running. Press CTRL+C to exit")

	// イベントハンドラの登録
	dg.AddHandler(voiceStateUpdate)

	// 無限ループでBotを実行し続ける
	select {}
}

// voiceStateUpdate：ボイスチャンネルの状態が更新されたときに呼ばれるイベント
func voiceStateUpdate(s *discordgo.Session, vsu *discordgo.VoiceStateUpdate) {
	if vsu == nil {
		log.Println("VoiceStateUpdate event is nil")
		return
	}

	// vsuがnilでないことが保証されているので、ここで変数を定義
	userID := vsu.UserID

	// チャンネルに参加した場合、現在の時間を記録
	if vsu.ChannelID == voiceChannelID && vsu.BeforeUpdate == nil { // ボイスチャンネルに参加
		userJoinTimes[userID] = time.Now()
		log.Printf("User %s has joined the voice channel at %v", userID, userJoinTimes[userID])
		return
	}

	// チャンネルを退出した場合の処理
	if vsu.BeforeUpdate != nil && vsu.ChannelID == "" {
		handleUserExit(s, userID)
	}
}

// handleUserExit：チャンネルを退出したときに呼び出される処理
func handleUserExit(s *discordgo.Session, userID string) {
	// ユーザーIDをキーに参加時刻を取得
	joinTime, ok := userJoinTimes[userID]
	if ok {
		// 滞在時間を計算
		duration := time.Since(joinTime)

		// メッセージを作成
		durationMessage := fmt.Sprintf("<@%s> Good job!! You stayed for %v.", userID, duration)

		// メッセージをDiscordの特定のチャンネルに送信
		_, err := s.ChannelMessageSend(textChannelID, durationMessage)
		if err != nil {
			log.Printf("Error sending message: %v", err)
		}

		// 参加時刻の削除
		delete(userJoinTimes, userID)
	} else {
		log.Printf("No join time found for user %s", userID)
	}
}
