package main

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/bwmarrin/discordgo"
	"google.golang.org/api/option"
)

// Firestore クライアントの初期化
func initFirestoreClient() (*firestore.Client, context.Context, error) {
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, "peachtech-mokumoku", option.WithCredentialsFile(CredentialsFile))
	if err != nil {
		return nil, nil, err
	}
	return client, ctx, nil
}

// Discord セッションの初期化
func initDiscordSession() (*discordgo.Session, error) {
	dg, err := discordgo.New("Bot " + DiscordToken)
	if err != nil {
		return nil, err
	}
	return dg, nil
}

// ユーザーデータを読み込む
func loadUserData(client *firestore.Client, ctx context.Context) ([]UserData, error) {
	return ReadUserProfiles(ctx, client)
}