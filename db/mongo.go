package db

import (
	"context"
	"log"
	"time"

	"github.com/itpu-student/s101_api/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	Client *mongo.Client
	DB     *mongo.Database
)

func Connect(ctx context.Context) {
	cctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(cctx, options.Client().ApplyURI(config.Cfg.MongoURI))
	if err != nil {
		log.Fatalf("mongo connect: %v", err)
	}
	if err := client.Ping(cctx, nil); err != nil {
		log.Fatalf("mongo ping: %v", err)
	}
	Client = client
	DB = client.Database(config.Cfg.MongoDB)
	log.Printf("connected to mongo: %s", config.Cfg.MongoDB)
}

func Disconnect(ctx context.Context) {
	if Client != nil {
		_ = Client.Disconnect(ctx)
	}
}

// Collection shortcuts.
func Users() *mongo.Collection         { return DB.Collection("users") }
func Admins() *mongo.Collection        { return DB.Collection("admins") }
func Categories() *mongo.Collection    { return DB.Collection("categories") }
func Places() *mongo.Collection        { return DB.Collection("places") }
func Reviews() *mongo.Collection       { return DB.Collection("reviews") }
func Bookmarks() *mongo.Collection     { return DB.Collection("bookmarks") }
func OTPCodes() *mongo.Collection      { return DB.Collection("otp_codes") }
func ClaimRequests() *mongo.Collection { return DB.Collection("claim_requests") }

func EnsureIndexes(ctx context.Context) {

	// Users
	mustIdx(ctx, Users(), mongo.IndexModel{
		Keys:    bson.D{{Key: "telegram_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	mustIdx(ctx, Users(), mongo.IndexModel{
		Keys:    bson.D{{Key: "username", Value: 1}},
		Options: options.Index().SetUnique(true).SetSparse(true),
	})

	// Admins
	mustIdx(ctx, Admins(), mongo.IndexModel{
		Keys:    bson.D{{Key: "username", Value: 1}},
		Options: options.Index().SetUnique(true),
	})

	// Categories
	mustIdx(ctx, Categories(), mongo.IndexModel{
		Keys:    bson.D{{Key: "slug", Value: 1}},
		Options: options.Index().SetUnique(true),
	})

	// OTP Codes
	mustIdx(ctx, OTPCodes(), mongo.IndexModel{Keys: bson.D{{Key: "code", Value: 1}}})
	mustIdx(ctx, OTPCodes(), mongo.IndexModel{Keys: bson.D{{Key: "telegram_id", Value: 1}}})
	mustIdx(ctx, OTPCodes(), mongo.IndexModel{Keys: bson.D{{Key: "expires_at", Value: 1}}})

	// Places
	mustIdx(ctx, Places(), mongo.IndexModel{
		Keys:    bson.D{{Key: "slug", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	mustIdx(ctx, Places(), mongo.IndexModel{Keys: bson.D{{Key: "atc_id", Value: 1}}})
	mustIdx(ctx, Places(), mongo.IndexModel{
		Keys: bson.D{{Key: "status", Value: 1}, {Key: "category_id", Value: 1}},
	})
	mustIdx(ctx, Places(), mongo.IndexModel{
		Keys: bson.D{
			{Key: "name", Value: "text"},
			{Key: "description.en", Value: "text"},
			{Key: "description.uz", Value: "text"},
		},
	})
	mustIdx(ctx, Places(), mongo.IndexModel{Keys: bson.D{{Key: "avg_rating", Value: -1}}})
	mustIdx(ctx, Places(), mongo.IndexModel{Keys: bson.D{{Key: "created_by", Value: 1}}})
	mustIdx(ctx, Places(), mongo.IndexModel{Keys: bson.D{{Key: "claimed_by", Value: 1}}})
	mustIdx(ctx, Places(), mongo.IndexModel{Keys: bson.D{{Key: "location", Value: "2dsphere"}}})

	// Reviews
	mustIdx(ctx, Reviews(), mongo.IndexModel{
		Keys: bson.D{{Key: "place_id", Value: 1}, {Key: "user_id", Value: 1}},
		Options: options.Index().SetUnique(true).
			SetPartialFilterExpression(bson.M{"latest": true}),
	})
	mustIdx(ctx, Reviews(), mongo.IndexModel{Keys: bson.D{{Key: "user_id", Value: 1}}})
	mustIdx(ctx, Reviews(), mongo.IndexModel{
		Keys: bson.D{
			{Key: "place_id", Value: 1},
			{Key: "user_id", Value: 1},
			{Key: "latest", Value: 1},
		},
	})

	// Bookmarks
	mustIdx(ctx, Bookmarks(), mongo.IndexModel{
		Keys:    bson.D{{Key: "user_id", Value: 1}, {Key: "place_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})

	// Claim Requests
	mustIdx(ctx, ClaimRequests(), mongo.IndexModel{
		Keys: bson.D{{Key: "place_id", Value: 1}, {Key: "status", Value: 1}},
	})
	mustIdx(ctx, ClaimRequests(), mongo.IndexModel{Keys: bson.D{{Key: "user_id", Value: 1}}})

	log.Println("indexes ensured")
}

func mustIdx(ctx context.Context, c *mongo.Collection, m mongo.IndexModel) {
	if _, err := c.Indexes().CreateOne(ctx, m); err != nil {
		log.Printf("index create on %s failed: %v", c.Name(), err)
	}
}
