package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"

	"cloud.google.com/go/firestore"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

var client *firestore.Client


type UserData struct {
	UserName          string
	WeeklyStayingTime int
}

var userDataList []UserData

func main() {
	// .envファイルから環境変数を読み込む
	err := godotenv.Load()
	if err != nil {
		log.Fatalf(".envファイルの読み込みに失敗しました: %v", err)
	}

	// Discordトークンの取得
	discordToken := os.Getenv("DISCORDTOKEN")
	if discordToken == "" {
		log.Fatal("環境変数DISCORDTOKENが設定されていません")
	}

	// Firestoreクライアントの設定、初期化
	ctx := context.Background()
	client, err = firestore.NewClient(ctx, "peachtech-mokumoku", option.WithCredentialsFile("./peachtech-mokumoku-91af9d3931c9.json"))
	if err != nil {
		log.Fatalf("Firestoreクライアントの初期化に失敗しました: %v", err)
	}
	// Firestoreクライアントの使用後、自動的にクローズ
	defer client.Close()

	// FirestoreからWeeklyTimeフィールドのみを読み込む
	ReadUserNameAndWeeklyStayingTime(ctx)

	// スライス内データをソートし、上位3名を表示する
	SortTop3Users()

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

	// 起動メッセージを表示
	fmt.Println("Bot is now running. Press CTRL+C to exit.")

}

// FirestoreからWeeklyTimeフィールドのみを取得する関数
func ReadUserNameAndWeeklyStayingTime(ctx context.Context) {
	// "user_profiles"コレクションから"UserName"と"WeeklyStayingTime"フィールドを選択して取得
	docRefs := client.Collection("user_profiles").Select("UserName", "WeeklyStayingTime").Documents(ctx)

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
func SortTop3Users() {
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