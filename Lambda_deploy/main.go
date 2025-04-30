package main

import (
	"log"
	//"github.com/aws/aws-lambda-go/lambda"
)

type UserData struct {
	UserID            string
	UserName          string
	WeeklyStayingTime int
}

func main() {
	// lambda.Start(handler)
	handler()
}

func handler() {
	// DevModeの設定を読み込む
	SetupDevMode()

	// Firestore クライアントの初期化
	client, ctx, err := initFirestoreClient()
	if err != nil {
		log.Fatalf("Firestoreクライアントの初期化に失敗しました: %v", err)
	}
	defer client.Close()

	// Firestore からユーザーデータを読み込む
	userDataList, err := loadUserData(client, ctx)
	if err != nil {
		log.Fatalf("ユーザーデータの取得に失敗しました: %v", err)
	}

	// Discord セッションの初期化
	dg, err := initDiscordSession()
	if err != nil {
		log.Fatalf("Discordセッションの作成に失敗しました: %v", err)
	}
	defer dg.Close()

	// Discord サーバーに接続
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