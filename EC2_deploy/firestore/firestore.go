// firestore.go
package firestore

import (
    "context"
    "fmt"
    "log"
    "time"

    "cloud.google.com/go/firestore"
    "github.com/bwmarrin/discordgo"
    "google.golang.org/api/option"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

type FirestoreService struct {
    Client *firestore.Client
}

func InitFirestore(credentialsFile string) (*firestore.Client, error) {
    ctx := context.Background()

    client, err := firestore.NewClient(ctx, "peachtech-mokumoku", option.WithCredentialsFile(credentialsFile))
    if err != nil {
        return nil, fmt.Errorf("failed to create Firestore client: %w", err)
    }

    log.Println("Firestore client initialized successfully")
    return client, nil
}

func NewFirestoreService(client *firestore.Client) *FirestoreService {
    return &FirestoreService{Client: client}
}

func (fs *FirestoreService) UpdateUserStayingTime(ctx context.Context, s *discordgo.Session, userID string, duration time.Duration) error {
    docRef := fs.Client.Collection("user_profiles").Doc(userID)

    docSnap, err := docRef.Get(ctx)
    if err != nil {
        if status.Code(err) != codes.NotFound {
            return fmt.Errorf("error retrieving document: %w", err)
        }
        // ドキュメントが存在しない場合の初期化
        docSnap = nil
    }

    var totalStayingTime int64 = 0
    var weeklyStayingTime int64 = 0

    if docSnap != nil && docSnap.Exists() {
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
