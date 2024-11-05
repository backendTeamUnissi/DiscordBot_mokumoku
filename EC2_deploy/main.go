package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

// グローバル変数の宣言！（初期化はmain関数内で行う）
var token string
var textChannelID string
var voiceChannelID string

// userJoinTimes：ユーザーIDをキーに参加時刻を記録するマップ
var userJoinTimes = make(map[string]time.Time)

// Firestoreクライアントをグローバルに宣言
var client *firestore.Client

func main() {
	// .envファイルから環境変数を読み込み
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Failed to load .env file: %v", err)
	}

	// 初期化
	token = os.Getenv("DISCORDTOKEN")
	if token == "" {
		log.Fatal("Discord token is not set. Please set the DISCORDTOKEN environment variable.")
	}

	textChannelID = os.Getenv("DISCORDTEXTCHANNELID")
	if textChannelID == "" {
		log.Fatal("Discord text channel ID is not set. Please set the DISCORDTEXTCHANNELID environment variable.")
	}

	voiceChannelID = os.Getenv("DISCORDVOICECHANNELID")
	if voiceChannelID == "" {
		log.Fatal("Discord voice channel ID is not set. Please set the DISCORDVOICECHANNELID environment variable.")
	}

	// Firestoreクライアントを初期化
	ctx := context.Background()
	client, err = firestore.NewClient(ctx, "peachtech-mokumoku", option.WithCredentialsFile("./peachtech-mokumoku-91af9d3931c9.json"))
	if err != nil {
		log.Fatalf("Failed to create Firestore client: %v", err)
	}
	defer client.Close()

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
	defer dg.Close()
	fmt.Println("Bot is now running. Press CTRL+C to exit.")

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

	userID := vsu.UserID

	// チャンネルに参加した場合、現在の時間を記録
	if vsu.ChannelID == voiceChannelID && vsu.BeforeUpdate == nil { // ボイスチャンネルに参加
		userJoinTimes[userID] = time.Now()
		joinTimeStr := userJoinTimes[userID].Format("2006-01-02 15:04:05")
		log.Printf("User %s joined the voice channel at %s", userID, joinTimeStr)
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

		// Discordユーザー情報の取得（ユーザー名など）
		user, err := s.User(userID)
		if err != nil {
			log.Printf("Error retrieving user information: %v", err)
			return
		}

		// 滞在時間をフォーマット
		durationStr := formatDuration(duration)

		// メッセージを作成してDiscordの特定のチャンネルに送信
		durationMessage := fmt.Sprintf("<@%s> お疲れ様でした！今回の滞在時間は %s です。", userID, durationStr)
		_, err = s.ChannelMessageSend(textChannelID, durationMessage)
		if err != nil {
			log.Printf("Error sending message: %v", err)
		}

		// Firestoreへのデータ送信
		ctx := context.Background()
		docRef := client.Collection("user_profiles").Doc(userID)

		// 既存のデータを取得
		docSnap, err := docRef.Get(ctx)
		if err != nil {
			log.Printf("Error retrieving document: %v", err)
			return
		}

		var totalStayingTime int64 = 0
		var weeklyStayingTime int64 = 0

		if docSnap.Exists() {
			data := docSnap.Data()
			// TotalStayingTimeを取得
			if val, ok := data["TotalStayingTime"].(int64); ok {
				totalStayingTime = val
			} else {
				log.Printf("TotalStayingTime is not of type int64")
			}
			// WeeklyStayingTimeを取得
			if val, ok := data["WeeklyStayingTime"].(int64); ok {
				weeklyStayingTime = val
			} else {
				log.Printf("WeeklyStayingTime is not of type int64")
			}
		}

		// 新しい滞在時間を加算
		durationSeconds := int64(duration.Seconds())
		totalStayingTime += durationSeconds
		weeklyStayingTime += durationSeconds

		// Firestoreにデータを書き込む
		_, err = docRef.Set(ctx, map[string]interface{}{
			"TotalStayingTime":  totalStayingTime,
			"UserID":            userID,
			"UserName":          user.Username,
			"UserRank":          0,
			"WeeklyStayingTime": weeklyStayingTime,
		})
		if err != nil {
			log.Printf("Error writing to Firestore: %v", err)
		}

		// ログに滞在時間を出力
		log.Printf("User %s's staying duration: %s", userID, durationStr)

		// 参加時刻の削除
		delete(userJoinTimes, userID)
	} else {
		log.Printf("Join time for user %s not found", userID)
	}
}

// formatDuration：滞在時間を時分秒の形式にフォーマット
func formatDuration(duration time.Duration) string {
	totalSeconds := int(duration.Seconds())
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60

	return fmt.Sprintf("%02d時間%02d分%02d秒", hours, minutes, seconds)
}
