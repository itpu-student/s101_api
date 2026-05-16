package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/itpu-student/s101_api/models"
	"github.com/itpu-student/s101_api/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func SeedTestUsers(ctx context.Context, count int) {
	now := time.Now().UTC()
	fmt.Printf("%-16s %-40s %s\n", "username", "id", "token")
	for n := 1; n <= count; n++ {
		username := fmt.Sprintf("tester%d", n)
		name := fmt.Sprintf("Tester %d", n)
		id := utils.NewUUIDv7()

		filter := bson.M{"username": username}
		update := bson.M{
			"$setOnInsert": bson.M{
				"_id":         id,
				"telegram_id": fmt.Sprintf("__test__%d", n),
				"phone":       fmt.Sprintf("+9999999%04d", n),
				"blocked":     false,
				"created_at":  now,
			},
			"$set": bson.M{
				"name":       name,
				"username":   username,
				"updated_at": now,
			},
		}
		res, err := Users().UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
		if err != nil {
			log.Printf("upsert tester%d: %v", n, err)
			continue
		}

		// fetch real id (may already exist)
		realID := id
		if res.UpsertedID == nil {
			var u models.User
			if err := Users().FindOne(ctx, filter).Decode(&u); err != nil {
				log.Printf("fetch tester%d: %v", n, err)
				continue
			}
			realID = u.ID
		}

		token, err := utils.IssueJWT(realID, utils.TypUser)
		if err != nil {
			log.Printf("jwt tester%d: %v", n, err)
			continue
		}
		fmt.Printf("%-16s %-40s %s\n", username, realID, token)
	}
}
