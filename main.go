package main

import (
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"

)

// グローバル変数の宣言！（初期化はmain関数内で行う）
var token string
var textChannelID string
var voiceChannelID string

// userJoinTimes：ユーザーIDをキーに参加時刻を記録するマップ
var userJoinTimes = make(map[string]time.Time)

func main() {
	// .envファイルから環境変数を読み込み
	init_env()

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
