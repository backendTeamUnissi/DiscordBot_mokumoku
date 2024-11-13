package discord

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"cloud.google.com/go/firestore"
	"DiscordBot_mokumoku/EC2_deploy/config"
    "DiscordBot_mokumoku/EC2_deploy/utils"

	"github.com/bwmarrin/discordgo"
)

type DiscordHandler struct {
	FirestoreClient     *firestore.Client
	TextChannelID       string
	VoiceChannelID      string
	UserJoinTimes       map[string]time.Time
	Mutex               sync.Mutex
}

func RegisterHandlers(s *discordgo.Session, fsClient *firestore.Client, cfg *config.Config) {
	handler := &DiscordHandler{
		FirestoreClient: fsClient,
		TextChannelID:   cfg.DiscordTextChannelID,
		VoiceChannelID:  cfg.DiscordVoiceChannelID,
		UserJoinTimes:   make(map[string]time.Time),
	}
	s.AddHandler(handler.VoiceStateUpdate)
}

func (h *DiscordHandler) VoiceStateUpdate(s *discordgo.Session, vsu *discordgo.VoiceStateUpdate) {
	if vsu == nil {
		log.Println("VoiceStateUpdate event is nil")
		return
	}

	userID := vsu.UserID

	h.Mutex.Lock()
	defer h.Mutex.Unlock()

	// チャンネルに参加した場合、現在の時間を記録
	if vsu.ChannelID == h.VoiceChannelID && vsu.BeforeUpdate == nil {
		h.UserJoinTimes[userID] = time.Now()
		joinTimeStr := h.UserJoinTimes[userID].Format("2006-01-02 15:04:05")
		log.Printf("User %s joined the voice channel at %s", userID, joinTimeStr)
		return
	}

	// チャンネルを退出した場合の処理
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
	err = h.updateFirestore(context.Background(), s, userID, duration)
	if err != nil {
		log.Printf("Error updating Firestore: %v", err)
	}

	log.Printf("User %s's staying duration: %s", userID, durationStr)

	// 参加時刻の削除
	delete(h.UserJoinTimes, userID)
}

func (h *DiscordHandler) updateFirestore(ctx context.Context, s *discordgo.Session, userID string, duration time.Duration) error {

	docRef := h.FirestoreClient.Collection("user_profiles").Doc(userID)

	docSnap, err := docRef.Get(ctx)
	if err != nil {
		return fmt.Errorf("error retrieving document: %w", err)
	}

	var totalStayingTime int64 = 0
	var weeklyStayingTime int64 = 0

	if docSnap.Exists() {
		data := docSnap.Data()
		if val, ok := data["TotalStayingTime"].(int64); ok {
			totalStayingTime = val
		} else {
			log.Println("TotalStayingTime is not of type int64")
		}
		if val, ok := data["WeeklyStayingTime"].(int64); ok {
			weeklyStayingTime = val
		} else {
			log.Println("WeeklyStayingTime is not of type int64")
		}
	}

	durationSeconds := int64(duration.Seconds())
	totalStayingTime += durationSeconds
	weeklyStayingTime += durationSeconds

	// Discordユーザー情報の取得
	user, err := s.User(userID)
	if err != nil {
		return fmt.Errorf("error retrieving user information: %w", err)
	}

	_, err = docRef.Set(ctx, map[string]interface{}{
		"TotalStayingTime":  totalStayingTime,
		"UserID":            userID,
		"UserName":          user.Username,
		"UserRank":          0,
		"WeeklyStayingTime": weeklyStayingTime,
	})
	if err != nil {
		return fmt.Errorf("error writing to Firestore: %w", err)
	}

	return nil
}
