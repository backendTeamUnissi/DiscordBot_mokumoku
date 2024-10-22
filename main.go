package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"

	// Firestore用のパッケージ
	"context"

	"cloud.google.com/go/firestore"
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
		log.Fatalf(".envファイルの読み込みに失敗しました: %v", err)
	}

	// 初期化
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

		// Discordユーザー情報の取得（ユーザー名など）
		user, err := s.User(userID)
		if err != nil {
			log.Printf("Error fetching user info: %v", err)
			return
		}

		// メッセージを作成してDiscordの特定のチャンネルに送信
		durationMessage := fmt.Sprintf("<@%s> Good job!! You stayed for %v.", userID, duration)
		_, err = s.ChannelMessageSend(textChannelID, durationMessage)
		if err != nil {
			log.Printf("Error sending message: %v", err)
		}

		// Firestoreへのデータ送信
		ctx := context.Background()
		docRef := client.Collection("user_profiles").Doc(userID) // コレクション名はusersとする
		if err != nil {
			log.Printf("Error creating Firestore document reference: %v", err)
		}
		// データ構造の定義とFirestoreへの書き込み
		_, err = docRef.Set(ctx, map[string]interface{}{
			"TotalStayingTime":  int64(duration.Seconds()), // 滞在時間（秒）
			"UserID":            userID,
			"UserName":          user.Username,
			"UserRank":          0,                         // 初期値または別のロジックで設定可能
			"WeeklyStayingTime": int64(duration.Seconds()), // 週間滞在時間（ここでは単純に今回の滞在時間）
		})
		if err != nil {
			log.Printf("Error writing to Firestore: %v", err)
		}

		// 参加時刻の削除
		delete(userJoinTimes, userID)
	} else {
		log.Printf("No join time found for user %s", userID)
	}
}
