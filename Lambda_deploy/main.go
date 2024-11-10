package main

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/firestore"
	"github.com/joho/godotenv"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

var client *firestore.Client

func main() {
	// .envファイルから環境変数を読み込む
	err := godotenv.Load()
	if err != nil {
		log.Fatalf(".envファイルの読み込みに失敗しました: %v", err)
	}

	// Firestoreクライアントの設定、初期化
	ctx := context.Background()
	client, err = firestore.NewClient(ctx, "peachtech-mokumoku", option.WithCredentialsFile("./peachtech-mokumoku-91af9d3931c9.json"))

	// Firestoreから全データを読み込む
    ReadAllFromFirestore(ctx)
}

// Firestoreから全データを読み込む関数
func ReadAllFromFirestore(ctx context.Context) {
    // 読み込むコレクションを指定
    docRefs := client.Collection("user_profiles").Documents(ctx)

    // ドキュメントを反復処理して取得
    for {
        docSnap, err := docRefs.Next()
        if err != nil {
            // エラーが発生した場合は終了
            if err == iterator.Done {
                break
            }
            log.Printf("Firestoreからのデータ取得エラー: %v", err)
            return
        }

  // ドキュメントのデータを取得し、コンソールに表示
        data := docSnap.Data() // Data() でマップとして取得
        if len(data) > 0 {
            fmt.Printf("Firestoreから読み取ったデータ: %v\n", data)

        }
    }
}