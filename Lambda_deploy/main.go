package main

import (
	"context"
	"fmt"
	"log"
	"sort"

	"cloud.google.com/go/firestore"
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

	// Firestoreクライアントの設定、初期化
	ctx := context.Background()
	client, err = firestore.NewClient(ctx, "peachtech-mokumoku", option.WithCredentialsFile("./peachtech-mokumoku-91af9d3931c9.json"))
	if err != nil {
		log.Fatalf("Firestoreクライアントの初期化に失敗しました: %v", err)
	}

	// FirestoreからWeeklyTimeフィールドのみを読み込む
	ReadUserNameAndWeeklyStayingTime(ctx)

	// スライス内データをソートし、上位3名を表示する
	SortTop3Users()
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