package main

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

// 秒を「○時間○分○秒」形式に変換する関数
func formatDuration(seconds int) string {
	duration := time.Duration(seconds) * time.Second
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	secs := int(duration.Seconds()) % 60
	return fmt.Sprintf("%d時間%d分%d秒", hours, minutes, secs)
}

// ユーザーの滞在時間を基にソートする共通関数
func sortUsersByStayingTime(userDataList []UserData) []UserData {
	validUsers := []UserData{}
	for _, user := range userDataList {
		if user.WeeklyStayingTime > 0 {
			validUsers = append(validUsers, user)
		}
	}

	// 滞在時間が多い順にソート
	sort.Slice(validUsers, func(i, j int) bool {
		return validUsers[i].WeeklyStayingTime > validUsers[j].WeeklyStayingTime
	})

	return validUsers
}

// Discordメッセージ送信用の関数（メイン関数から呼び出される）
func SendMessages(s *discordgo.Session, channelID string, userDataList []UserData) {
	sendNormalMessage(s, channelID, userDataList)
	sendEmbedMessage(s, channelID, userDataList)
}

// 通常のテキストメッセージを送信
func sendNormalMessage(s *discordgo.Session, channelID string, userDataList []UserData) {
	validUsers := sortUsersByStayingTime(userDataList)

	mentionMessage := ""
	for i := 0; i < 3 && i < len(validUsers); i++ {
		mentionMessage += fmt.Sprintf("<@%s> ", validUsers[i].UserID)
	}
	if mentionMessage == "" {
		return
	}
	_, err := s.ChannelMessageSend(channelID, mentionMessage)
	if err != nil {
		fmt.Println("Error sending normal message:", err)
	}
}

// Embedメッセージを送信
func sendEmbedMessage(s *discordgo.Session, channelID string, userDataList []UserData) {
	validUsers := sortUsersByStayingTime(userDataList)

	rankedNum := len(validUsers)
	maxRankNum := 3
	showRankNum := rankedNum
	// ユーザーが3以上の時は、テキストをトップ３で固定
	if showRankNum > maxRankNum {
		showRankNum = maxRankNum
	}

	title := fmt.Sprintf("🔥今週の滞在時間トップ%d🔥", showRankNum)
	var descriptionBuilder strings.Builder

	if rankedNum == 0 {
		title = "今週の滞在者なし😢"
		descriptionBuilder.WriteString("今週はもくもくしていませんでした…\n")
	} else {
		descriptionBuilder.WriteString("今週のもくもくを頑張ったユーザーはこちら！\n")
	}

	for i := 0; i < maxRankNum; i++ {
		if i < rankedNum {
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