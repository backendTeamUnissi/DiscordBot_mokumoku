// main.go
package main

import (
    "fmt"
    "log"

    // パッケージのインポート
    "DiscordBot_mokumoku/EC2_deploy/config"
    "DiscordBot_mokumoku/EC2_deploy/discord"
    "DiscordBot_mokumoku/EC2_deploy/firestore"
)

func main() {
    // 環境変数の読み込み
    cfg, err := config.LoadConfig()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    // Firestoreクライアントの初期化
    fsClient, err := firestore.InitFirestore(cfg.FirestoreCredentials)
    if err != nil {
        log.Fatalf("Failed to initialize Firestore: %v", err)
    }
    defer fsClient.Close()

    // Discordセッションの初期化
    dg, err := discord.InitDiscord(cfg.DiscordToken)
    if err != nil {
        log.Fatalf("Failed to initialize Discord: %v", err)
    }
    defer dg.Close()

    // イベントハンドラの登録
	discord.RegisterHandlers(dg, fsClient, cfg)


    // Botを起動
    err = dg.Open()
    if err != nil {
        log.Fatalf("Error opening Discord connection: %v", err)
    }
    defer dg.Close()

    fmt.Println("Bot is now running. Press CTRL+C to exit.")

    // 無限ループで実行
    select {}
}
