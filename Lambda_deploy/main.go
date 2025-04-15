package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/bwmarrin/discordgo"
	"google.golang.org/api/option"
)

type UserData struct {
	UserID            string
	UserName          string
	WeeklyStayingTime int
}

// グローバル変数の宣言！（初期化はmain関数内で行う）
var userDataList []UserData
var err error

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
	client, err = firestore.NewClient(ctx, "peachtech-mokumoku", option.WithCredentialsFile("./peachtech-mokumoku-91af9d3931c9.json"))
	if err != nil {
		log.Fatalf("Firestoreクライアントの初期化に失敗しました: %v", err)
	}
	// リソースの解放
	defer client.Close()

	// Firestoreからユーザーデータを取得する
	ReadUserProfiles(ctx, client)

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

	// 上位3名の情報をEmbed/通常のメッセージ形式に組み立てて送信
	SendMessages(dg, TextChannelID, userDataList)

	// WeeklyStayingTimeをリセットする
	ResetWeeklyStayingTime(ctx)
	
}

// すでに取得してあるユーザーデータのスライスを用い、WeeklyStayingTimeをリセットする関数
func ResetWeeklyStayingTime(ctx context.Context) {
	// 各ユーザーのWeeklyStayingTimeをリセット
	for _, userData := range userDataList {
		_, err := client.Collection(CollectionName).Doc(userData.UserID).Update(ctx, []firestore.Update{
			{Path: "WeeklyStayingTime", Value: 0}, // WeeklyStayingTimeを0にリセット
		})
		if err != nil {
			log.Printf("ユーザー %s のWeeklyStayingTimeのリセットに失敗しました: %v", userData.UserID, err)
		} else {
			log.Printf("ユーザー %s のWeeklyStayingTimeをリセットしました", userData.UserID)
		}
	}

	fmt.Println("\nWeeklyStayingTimeがリセットされました。")
}