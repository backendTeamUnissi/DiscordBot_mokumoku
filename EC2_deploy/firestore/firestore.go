package firestore

import (
	"context"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/option"
)

func InitFirestore(credentialsFile string) (*firestore.Client, error) {
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, "peachtech-mokumoku", option.WithCredentialsFile(credentialsFile))
	if err != nil {
		return nil, err
	}
	return client, nil
}
