package main

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/firestore"
)

// すでに取得してあるユーザーデータのスライスを用い、WeeklyStayingTimeをリセットする関数
func ResetWeeklyStayingTime(ctx context.Context, client *firestore.Client, userDataList []UserData) {
	for _, userData := range userDataList {
		_, err := client.Collection(CollectionName).Doc(userData.UserID).Update(ctx, []firestore.Update{
			{Path: "WeeklyStayingTime", Value: 0},
		})
		if err != nil {
			log.Printf("ユーザー %s のWeeklyStayingTimeのリセットに失敗しました: %v", userData.UserID, err)
		} else {
			log.Printf("ユーザー %s のWeeklyStayingTimeをリセットしました", userData.UserID)
		}
	}

	fmt.Println("\nWeeklyStayingTimeがリセットされました。")
}