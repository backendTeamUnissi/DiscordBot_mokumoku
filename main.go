package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

// userJoinTimes：ユーザーIDをキーに参加時刻を記録するマップ
// キー: ユーザーID (string), 値: 参加した時刻 (time.Time)
var userJoinTimes = make(map[string]time.Time)

func main() {
	// Loadnv：.envファイルから環境変数を取得
	// 戻り値: error（読み込みエラー時）
	err := godotenv.Load()
	if err != nil {
		log.Fatalf(".envファイルの読み込みに失敗しました: %v", err)
	}

	// Getenv：環境変数からDiscordBotのトークンを取得
	// 戻り値: string (Discord認証トークン)
	token := os.Getenv("DISCORDTOKEN")
	if token == "" {
		log.Fatal("Discordトークンが設定されていません。環境変数DISCORDTOKENを設定してください。")
	}

	// Getenv：環境変数からチャンネルIDを取得
	textChannelID := os.Getenv("DISCORDTEXTCHANNELID")
	if textChannelID == "" {
		log.Fatal("DiscordチャンネルIDが設定されていません。環境変数DISCORDCHANNELIDを設定してください。")
	}

	// Getenv: 環境変数からボイスチャンネルIDを取得
	voiceChannelID := os.Getenv("DISCORDVOICECHANNELID")
	if voiceChannelID == "" {
		log.Fatal("DiscordボイスチャンネルIDが設定されていません。環境変数DISCORDVOICECHANNELIDを設定してください。")
	}

	// discordgo.New：DiscordAPIに接続するためのセクションを作成
	// token: Discordボットのトークン
	// dg: Discordセッションのインスタンス
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("Error creating Discord session: %v", err)
	}

	// Open()：BotをDiscordサーバーに接続し、起動する
	// 戻り値: error (接続エラー時)
	err = dg.Open()
	if err != nil {
		log.Fatalf("Error opening connection: %v", err)
	}

	// Botがシャットダウンされたときにセッションを閉じる
	defer dg.Close()
	fmt.Println("Bot is now running. Press CTRL+C to exit")

    // ボイスチャンネルの入退出を監視
	// s: Discordセッション, vsu: ボイスステートの更新情報（ユーザーのボイスチャンネルの状態）
    dg.AddHandler(func(s *discordgo.Session, vsu *discordgo.VoiceStateUpdate) {
        voiceStateUpdate(s, vsu, textChannelID, voiceChannelID) // 引数を適切に渡す
    })

	// 無限ループでBotを実行し続ける
	select {}
}

// voiceStateUpdate：ボイスチャンネルの状態が更新されたときに呼ばれるイベントハンド￥
func voiceStateUpdate(s *discordgo.Session, vsu *discordgo.VoiceStateUpdate, textChannelID, voiceChannelID string) {
    if vsu == nil {
        log.Println("VoiceStateUpdate event is nil")
        return
    }

	// vsuがnilでないことが保証されているので、ここで変数を定義
    userID := vsu.UserID

    // チャンネルに参加した場合、現在の時間を記録
	// vsu.ChannelID: チャンネルのID, vsu.BeforeUpdate: 以前の状態
    if vsu.ChannelID == voiceChannelID && vsu.BeforeUpdate == nil { // ボイスチャンネルに参加
        userJoinTimes[userID] = time.Now()
        log.Printf("User %s has joined the voice channel at %v", userID, userJoinTimes[userID])
        return
    }

	// チャンネルを退出した場合の処理
    if vsu.BeforeUpdate != nil && vsu.ChannelID == "" {
        handleUserExit(s, userID, textChannelID) // textChannelIDを引数に渡す
    }
}

// ユーザーが退出したときの処理を行う関数
// handleUserExit：チャンネルを退出したときに呼び出される処理
// s: Discordセッション, userID: 退出したユーザーのID, textChannelID: メッセージを送信するチャンネルID
func handleUserExit(s *discordgo.Session, userID, textChannelID string) {
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