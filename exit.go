package main

import (

	"fmt"
	"log"
	"time"
	"github.com/bwmarrin/discordgo"
)
// handleUserExit：チャンネルを退出したときに呼び出される処理
func handleUserExit(s *discordgo.Session, userID string) {
	// ユーザーIDをキーに参加時刻を取得
	joinTime, ok := userJoinTimes[userID]
	if ok {
		// 滞在時間を計算
		duration := time.Since(joinTime)

		// メッセージを作成
		durationMessage := fmt.Sprintf("<@%s> Good job!! You stayed for %v.", userID, duration)

		// メッセージをDiscordの特定のチャンネルに送信
		_, err := s.ChannelMessageSend(textChannelID, durationMessage)
		if err != nil {
			log.Printf("Error sending message: %v", err)
		}

		// 参加時刻の削除
		delete(userJoinTimes, userID)
	} else {
		log.Printf("No join time found for user %s", userID)
	}
}
