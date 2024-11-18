// handlers.go
package discord

import (
    "context"
    "fmt"
    "log"
    "sync"
    "time"

    "DiscordBot_mokumoku/EC2_deploy/firestore"
    "DiscordBot_mokumoku/EC2_deploy/utils"
    "github.com/bwmarrin/discordgo"
)

type DiscordHandler struct {
    FirestoreService *firestore.FirestoreService
    TextChannelID    string
    VoiceChannelID   string
    UserJoinTimes    map[string]time.Time
    Mutex            sync.Mutex
}

func (h *DiscordHandler) VoiceStateUpdate(s *discordgo.Session, vsu *discordgo.VoiceStateUpdate) {
    if vsu == nil {
        log.Println("VoiceStateUpdate event is nil")
        return
    }

    userID := vsu.UserID

    h.Mutex.Lock()
    defer h.Mutex.Unlock()

    // ユーザーがボイスチャンネルに参加した場合
    if vsu.ChannelID == h.VoiceChannelID && vsu.BeforeUpdate == nil {
        h.UserJoinTimes[userID] = time.Now()
        joinTimeStr := h.UserJoinTimes[userID].Format("2006-01-02 15:04:05")
        log.Printf("User %s joined the voice channel at %s", userID, joinTimeStr)
        return
    }

    // ユーザーがボイスチャンネルから退出した場合
    if vsu.BeforeUpdate != nil && vsu.ChannelID == "" {
        h.handleUserExit(s, userID)
    }
}

func (h *DiscordHandler) handleUserExit(s *discordgo.Session, userID string) {
    joinTime, ok := h.UserJoinTimes[userID]
    if !ok {
        log.Printf("Join time for user %s not found", userID)
        return
    }

    duration := time.Since(joinTime)
    durationStr := utils.FormatDuration(duration)

    // メッセージを送信
    durationMessage := fmt.Sprintf("<@%s> お疲れ様でした！今回の滞在時間は %s です。", userID, durationStr)
    _, err := s.ChannelMessageSend(h.TextChannelID, durationMessage)
    if err != nil {
        log.Printf("Error sending message: %v", err)
    }

    // Firestoreにデータを送信
    err = h.FirestoreService.UpdateUserStayingTime(context.Background(), s, userID, duration)
    if err != nil {
        log.Printf("Error updating Firestore: %v", err)
    }

    log.Printf("User %s's staying duration: %s", userID, durationStr)

    // 参加時刻の削除
    delete(h.UserJoinTimes, userID)
}
