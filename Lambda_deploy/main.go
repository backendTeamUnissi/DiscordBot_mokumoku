package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/bwmarrin/discordgo"
	"google.golang.org/api/option"
	//"github.com/aws/aws-lambda-go/lambda"
)

type UserData struct {
	UserID            string
	UserName          string
	WeeklyStayingTime int
}

// グローバル変数の宣言！（初期化はmain関数内で行う）
var userDataList []UserData

// 秒を「○時間○分○秒」形式に変換する関数
func formatDuration(seconds int) string {
	duration := time.Duration(seconds) * time.Second
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	secs := int(duration.Seconds()) % 60
	return fmt.Sprintf("%d時間%d分%d秒", hours, minutes, secs)
}

func main() {
	// lambda.Start(handler)
	handler()
}

func handler() {
	// DevModeの設定を読み込む
	SetupDevMode()

	// Firestoreクライアントの設定、初期化
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, "peachtech-mokumoku", option.WithCredentialsFile(CredentialsFile))
	if err != nil {
		log.Fatalf("Firestoreクライアントの初期化に失敗しました: %v", err)
	}
	// リソースの解放
	defer client.Close()

	// Firestoreからユーザーデータを取得する
	userDataList, err = ReadUserProfiles(ctx, client)
	if err != nil {
		log.Fatalf("ユーザーデータの取得に失敗しました: %v", err)
	}

	// Discord APIに接続
	dg, err := discordgo.New("Bot " + DiscordToken)
	if err != nil {
		log.Fatalf("Discordセッションの作成に失敗しました: %v", err)
	}
	// Botを起動し、Discordサーバーに接続
	err = dg.Open()
	if err != nil {
		log.Fatalf("Discordサーバーへの接続に失敗しました: %v", err)
	}
	// Discordセッションの使用後、自動的にクローズ
	defer dg.Close()

	// メッセージ送信
	SendMessages(dg, TextChannelID, userDataList)

	// WeeklyStayingTimeのリセット
	ResetWeeklyStayingTime(ctx, client, userDataList)
}