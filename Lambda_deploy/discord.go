package main

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// Discordメッセージ送信用の関数（メイン関数から呼び出される）
func SendMessages(s *discordgo.Session, channelID string, userDataList []UserData) {
	sendNormalMessage(s, channelID, userDataList)
	sendEmbedMessage(s, channelID, userDataList)
}

// 通常のテキストメッセージを送信
func sendNormalMessage(s *discordgo.Session, channelID string, userDataList []UserData) {
	message := ""
	for i := 0; i < 3 && i < len(userDataList); i++ {
		if userDataList[i].WeeklyStayingTime > 0 {
			message += fmt.Sprintf("<@%s> ", userDataList[i].UserID)
		}
	}
	if message == "" {
		return
	}
	_, err := s.ChannelMessageSend(channelID, message)
	if err != nil {
		fmt.Println("Error sending normal message:", err)
	}
}

// Embedメッセージを送信
func sendEmbedMessage(s *discordgo.Session, channelID string, userDataList []UserData) {
	validUsers := []UserData{}
	for _, user := range userDataList {
		if user.WeeklyStayingTime > 0 {
			validUsers = append(validUsers, user)
		}
	}

	rankCount := len(validUsers)
	title := fmt.Sprintf("🔥今週の滞在時間トップ%d🔥", rankCount)

	var descriptionBuilder strings.Builder
	if rankCount == 0 {
		title = "今週の滞在者なし😢"
		descriptionBuilder.WriteString("今週はもくもくしていませんでした…\n")
	} else {
		descriptionBuilder.WriteString("今週のもくもくを頑張ったユーザーはこちら！\n")
	}

	for i := 0; i < 3; i++ {
		if i < rankCount {
			userID := validUsers[i].UserID
			stayingTime := formatDuration(validUsers[i].WeeklyStayingTime)

			if i == 0 {
				descriptionBuilder.WriteString("\n")
			}

			descriptionBuilder.WriteString(fmt.Sprintf("**%d位:** <@%s>\n**滞在時間:** %s\n", i+1, userID, stayingTime))
		} else {
			descriptionBuilder.WriteString(fmt.Sprintf("**%d位:** ---\n**滞在時間:** ---\n", i+1))
		}
	}

	embed := &discordgo.MessageEmbed{
		Title:       title,
		Description: descriptionBuilder.String(),
		Color:       0x00ff00,
	}

	_, err := s.ChannelMessageSendEmbed(channelID, embed)
	if err != nil {
		fmt.Println("Error sending embed message:", err)
	}
}