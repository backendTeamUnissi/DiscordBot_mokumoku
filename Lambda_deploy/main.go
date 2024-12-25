package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"

	// "strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// DevModeを定義（開発モードを有効にする）
const DevMode = true

type UserData struct {
	UserName          string
	WeeklyStayingTime int
}

// グローバル変数の宣言！（初期化はmain関数内で行う）
var discordToken string
var textChannelID string
var userDataList []UserData
var client *firestore.Client
var err error

func loadEnv() {
    // 環境に応じた設定ファイルを読み込む
    var envFile string
    if DevMode {
        envFile = ".env.dev"  // 開発環境用
    } else {
        envFile = ".env.prod" // 本番環境用
    }

    // .envファイルから環境変数を読み込む
    err := godotenv.Load(envFile)
    if err != nil {
        log.Fatalf("%sの読み込みに失敗しました: %v", envFile, err)
    }

    // Discordトークンの取得
    discordToken = os.Getenv("DISCORDTOKEN")
    if discordToken == "" {
        log.Fatal("環境変数DISCORDTOKENが設定されていません")
    }

    // メッセージを送信するチャンネルIDの指定
    textChannelID = os.Getenv("DISCORDTEXTCHANNELID")
    if textChannelID == "" {
        log.Fatal("環境変数DISCORDTEXTCHANNELIDが設定されていません")
    }
}

func printMode() {
    if DevMode {
        fmt.Println("現在、開発モードで実行中です。")
    } else {
        fmt.Println("現在、本番モードで実行中です。")
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
	printMode()
	// lambda.Start(handler)
	handler()
}

func handler() {
	//環境変数の読み込み
	loadEnv()

	// Firestoreクライアントの設定、初期化
	ctx := context.Background()
	client, err = firestore.NewClient(ctx, "peachtech-mokumoku", option.WithCredentialsFile("./peachtech-mokumoku-91af9d3931c9.json"))
	if err != nil {
		log.Fatalf("Firestoreクライアントの初期化に失敗しました: %v", err)
	}
	// リソースの解放
	defer client.Close()

	// FirestoreからWeeklyTimeフィールドのみを読み込む
	ReadUserProfiles(ctx)

	// スライス内データをソートし、上位3名を表示する
	SortTopUsers()

	// Discord APIに接続
	dg, err := discordgo.New("Bot " + discordToken)
	if err != nil {
		log.Fatalf("Discordセッションの作成に失敗しました: %v", err)
	}
    // Discordセッションの使用後、自動的にクローズ
	defer dg.Close()

	// Botを起動し、Discordサーバーに接続
	err = dg.Open()
	if err != nil {
		log.Fatalf("Discordサーバーへの接続に失敗しました: %v", err)
	}

	// 上位3名の情報をメッセージとして組み立てて送信
	message := BuildTopUsersEmbed()

	// Discordチャンネルへのメッセージ送信
	_, err = dg.ChannelMessageSendEmbed(textChannelID, message)
	if err != nil {
		log.Printf("Discordへのメッセージ送信に失敗しました: %v", err)
	}

	// WeeklyStayingTimeをリセットする
	ResetWeeklyStayingTime(ctx)
	
}

// FirestoreからWeeklyTimeフィールドのみを取得する関数
func ReadUserProfiles(ctx context.Context) {

	// コレクション名をDevModeに基づいて切り替え
	var collectionName string
	if DevMode {
		collectionName = "test_profiles" // 開発環境用コレクション
	} else {
		collectionName = "user_profiles" // 本番環境用コレクション
	}

	// "user_profiles"コレクションから"UserName"と"WeeklyStayingTime"フィールドを選択して取得
	docRefs := client.Collection(collectionName).Select("UserName", "WeeklyStayingTime").Documents(ctx)

	// ドキュメントを反復処理して取得
	for {
		docSnap, err := docRefs.Next()
		if err != nil {
			// エラー（もうデータない）が発生した場合は終了
			if err == iterator.Done {
				break
			}
			log.Printf("Firestoreからのデータ取得エラー: %v", err)
			return
		}

		// ドキュメントのデータを取得し、コンソールに表示
		data := docSnap.Data() // Data() でマップとして取得
		if len(data) > 0 {
			// Firestoreから読み取ったデータを表示
			fmt.Printf("Firestoreから読み取ったデータ: %v\n", data)

			// UserData型にデータを格納
			userData := UserData{
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
}

    // スライスをソートし、上位3名を表示する関数
func SortTopUsers() {
	// WeeklyStayingTimeで降順にソート
	sort.Slice(userDataList, func(i, j int) bool {
		return userDataList[i].WeeklyStayingTime > userDataList[j].WeeklyStayingTime
	})

	// 上位3名を表示
	fmt.Println("\n上位3名のユーザー:")
	for i := 0; i < 3 && i < len(userDataList); i++ {
		fmt.Printf("%d位: %s - %d分\n", i+1, userDataList[i].UserName, userDataList[i].WeeklyStayingTime)
	}
}

// 上位3名の情報をメッセージとして構築
func BuildTopUsersEmbed() *discordgo.MessageEmbed {
	// Embedの基本情報を設定
	embed := &discordgo.MessageEmbed{
		Title:       "🔥今週の滞在時間トップ3🔥",
		Description: "今週のもくもくを頑張ったユーザーはこちら！",
		Color:       0xff0000, // グリーン (必要に応じて変更可能)
		Fields:      []*discordgo.MessageEmbedField{},
	}

	// 上位3名の情報をEmbedに追加
	for i := 0; i < 3 && i < len(userDataList); i++ {
		fieldName := fmt.Sprintf("%d位: %s", i+1, userDataList[i].UserName)
		fieldValue := fmt.Sprintf("滞在時間: %s", formatDuration(userDataList[i].WeeklyStayingTime))
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   fieldName,
			Value:  fieldValue,
			Inline: false, // 必要に応じてtrueに変更
		})
	}

	return embed
}

// Firestore内の全ユーザーのWeeklyStayingTimeを0にリセットする関数
func ResetWeeklyStayingTime(ctx context.Context) {
	var collectionName string

	// DevMode確認
	if DevMode {
		// 開発環境ではtest_profilesコレクションを対象にする
		collectionName = "test_profiles"
	} else {
		// 本番環境ではuser_profilesコレクションを対象にする
		collectionName = "user_profiles"
	}

	// 選択したコレクションの全ドキュメントを取得
	docRefs := client.Collection(collectionName).Documents(ctx)


	// 取得したドキュメントを1つずつ処理（ループ）
	for {
		docSnap, err := docRefs.Next()
		if err != nil {
			// 全てのデータを処理し終えた場合、ループを終了
			if err == iterator.Done {
				break
			}
			log.Printf("Firestoreからのデータ取得エラー: %v", err)
			return
		}

		// ドキュメントIDを取得
		docID := docSnap.Ref.ID

		// WeeklyStayingTimeフィールドを0に書き換える
		//  書き換えた値をFirestoreに保存する
		_, err = client.Collection(collectionName).Doc(docID).Update(ctx, []firestore.Update{
			{Path: "WeeklyStayingTime", Value: 0},
		})
		if err != nil {
			log.Printf("WeeklyStayingTimeのリセットに失敗しました (ID: %s): %v", docID, err)
		} else {
			log.Printf("WeeklyStayingTimeをリセットしました (ID: %s)", docID)
		}
	}
	log.Println("全ユーザーのWeeklyStayingTimeをリセットしました")
}