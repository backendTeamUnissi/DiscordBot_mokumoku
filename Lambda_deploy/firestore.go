package main

import (
	"context"
	"log"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

const (
	ProjectID       = "peachtech-mokumoku" // プロジェクトID
	CredentialsFile = "./peachtech-mokumoku-91af9d3931c9.json" // サービスアカウントの認証ファイル
)

// Firestoreクライアントの初期化
func InitFirestore() (*firestore.Client, error) {
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, ProjectID, option.WithCredentialsFile(CredentialsFile))
	if err != nil {
		log.Fatalf("Firestoreクライアントの初期化に失敗しました: %v", err)
		return nil, err
	}
	return client, nil
}

// Firestoreからユーザーデータを取得する関数
func ReadUserProfiles(ctx context.Context, client *firestore.Client) ([]UserData, error) {
	var userDataList []UserData
	// 指定したコレクションから"UserID","UserName","WeeklyStayingTime"フィールドを取得
	docRefs := client.Collection(CollectionName).Select("UserID", "UserName", "WeeklyStayingTime").Documents(ctx)

	// ドキュメントを反復処理して取得
	for {
		docSnap, err := docRefs.Next()
		if err != nil {
			// データの終わりを検知し、ループを終了
			if err == iterator.Done {
				break
			}
			log.Printf("Firestoreからのデータ取得エラー: %v", err)
			return nil, err
		}

		// 取得したフィールドをマップ形式で取得し、構造体に変換してスライスに保存
		data := docSnap.Data() // Data() でマップとして取得
		if len(data) > 0 {
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
	return userDataList, nil
}