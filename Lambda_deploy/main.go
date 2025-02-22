package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// DevModeを定義（テスト時は開発モードを有効にする）
const DevMode = true

type UserData struct {
	UserID            string
	UserName          string
	WeeklyStayingTime int
}

// グローバル変数の宣言！（初期化はmain関数内で行う）
var discordToken string
var textChannelID string
var userDataList []UserData
var client *firestore.Client
var err error
var collectionName string

// DevModeの設定を集約する関数
func setupDevMode() {
    // 環境モードに応じた設定を行う
	var envFile string
    if DevMode {
        envFile = ".env.dev"  // 開発環境用
		collectionName = "test_profiles" // 開発用コレクション
		fmt.Println("現在、開発モードで実行中です。")
    } else {
        envFile = ".env.prod" // 本番環境用
		collectionName = "user_profiles" // 本番用コレクション
		fmt.Println("現在、本番モードで実行中です。")
    }
    // 環境変数をロード
    err := godotenv.Load(envFile)
    if err != nil {
        log.Fatalf("%sの読み込みに失敗しました: %v", envFile, err)
    }

    // Discordトークンの取得
    discordToken = os.Getenv("DISCORDTOKEN")
    if discordToken == "" {
        log.Fatal("環境変数DISCORDTOKENが設定されていません")
    }

    // 環境モードに応じたメッセージ送信先のチャンネルID指定
    textChannelID = os.Getenv("DISCORDTEXTCHANNELID")
    if textChannelID == "" {
        log.Fatal("環境変数DISCORDTEXTCHANNELIDが設定されていません")
    }
}

// 秒を「○時間○分○秒」形式に変換する関数
func formatDuration(seconds int) string {
	duration := time.Duration(seconds) * time.Second
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	secs := int(duration.Seconds()) % 60
	return fmt.Sprintf("%d時間%d分%d秒", hours, minutes, secs)
}

func main() {
	lambda.Start(handler)
	handler()
}

func handler() {
	// DevModeの設定を読み込む
	setupDevMode()

	// Firestoreクライアントの設定、初期化
	ctx := context.Background()
	client, err = firestore.NewClient(ctx, "peachtech-mokumoku", option.WithCredentialsFile("./peachtech-mokumoku-91af9d3931c9.json"))
	if err != nil {
		log.Fatalf("Firestoreクライアントの初期化に失敗しました: %v", err)
	}
	// リソースの解放
	defer client.Close()

	// Firestoreからユーザーデータを取得する
	ReadUserProfiles(ctx)

	// Discord APIに接続
	dg, err := discordgo.New("Bot " + discordToken)
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
	sendMessages(dg, textChannelID)

	// WeeklyStayingTimeをリセットする
	ResetWeeklyStayingTime(ctx)
	
}

func sendMessages(s *discordgo.Session, channelID string) {
	sendNormalMessage(s, channelID, userDataList)
	sendEmbedMessage(s, channelID, userDataList)
}

// Embedメッセージを送信する関数
func sendEmbedMessage(s *discordgo.Session, channelID string, userDataList []UserData) {
	// Embedメッセージを作成
	embed := &discordgo.MessageEmbed{
		Title:       "🔥今週の滞在時間トップ3🔥",
		Description: "今週のもくもくを頑張ったユーザーはこちら！\n", // ここで改行を入れる
		Color:       0x00ff00,
	}

	// 上位3名のユーザー情報をEmbedのDescriptionに追加
	for i := 0; i < 3 && i < len(userDataList); i++ {
		// ユーザーIDと滞在時間を取得
		userID := userDataList[i].UserID
		stayingTime := formatDuration(userDataList[i].WeeklyStayingTime)
		// ユーザー情報（順位、ユーザーID、滞在時間）をEmbedのDescriptionに追加
		embed.Description += fmt.Sprintf("%d位: <@%s>\n滞在時間: %s\n", i+1, userID, stayingTime)
	}

	// 環境モードに応じたDiscordチャンネルへEmbedメッセージを送信
	_, err := s.ChannelMessageSendEmbed(channelID, embed)
	if err != nil {
		fmt.Println("Error sending embed message:", err)
		return
	}
}


// 通常のテキストメッセージを送信する関数
func sendNormalMessage(s *discordgo.Session, channelID string, userDataList []UserData) {
	message := ""
	// 上位3名のユーザーをメンション形式で、1行で組み立て
	for i := 0; i < 3 && i < len(userDataList); i++ {
		userID := userDataList[i].UserID
		message += fmt.Sprintf("<@%s> ", userID)
	}

	// メンション付きのテキストメッセージを送信
	_, err := s.ChannelMessageSend(channelID, message)
	if err != nil {
		fmt.Println("Error sending normal message:", err)
		return
	}
}

// Firestoreからユーザーデータを取得する関数
func ReadUserProfiles(ctx context.Context) {
	// 指定したコレクションから"UserID","UserName","WeeklyStayingTime"フィールドを取得
	docRefs := client.Collection(collectionName).Select("UserID", "UserName", "WeeklyStayingTime").Documents(ctx)

	// ドキュメントを反復処理して取得
	for {
		docSnap, err := docRefs.Next()
		if err != nil {
			// データの終わりを検知し、ループを終了
			if err == iterator.Done {
				break
			}
			log.Printf("Firestoreからのデータ取得エラー: %v", err)
			return
		}

		// 取得したフィールドをマップ形式で取得し、構造体に変換してスライスに保存
		data := docSnap.Data() // Data() でマップとして取得
		if len(data) > 0 {
			// Firestoreから読み取ったデータを表示
			fmt.Printf("Firestoreから読み取ったデータ: %v\n", data)

			// UserData型にデータを格納
			userData := UserData{
				UserID:            data["UserID"].(string), 
				UserName:          data["UserName"].(string),
				WeeklyStayingTime: int(data["WeeklyStayingTime"].(int64)), 
			}

			// ユーザーデータをスライスに追加
			userDataList = append(userDataList, userData)
		}
	}
	// スライスに格納された全ユーザーデータのデバッグ表示
	fmt.Println("\nFirestoreから取得した全ユーザーデータ:")
	for _, user := range userDataList {
		fmt.Printf("UserName: %s, WeeklyStayingTime: %d\n", user.UserName, user.WeeklyStayingTime)
	}

        // userDataListの降順ソート
		sort.Slice(userDataList, func(i, j int) bool {
		return userDataList[i].WeeklyStayingTime > userDataList[j].WeeklyStayingTime
	})
}

// すでに取得してあるユーザーデータのスライスを用い、WeeklyStayingTimeをリセットする関数
func ResetWeeklyStayingTime(ctx context.Context) {
	// 各ユーザーのWeeklyStayingTimeをリセット
	for _, userData := range userDataList {
		_, err := client.Collection(collectionName).Doc(userData.UserID).Update(ctx, []firestore.Update{
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