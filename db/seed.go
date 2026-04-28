package db

import (
	"context"
	"log"
	"time"

	"github.com/itpu-student/s101_api/config"
	"github.com/itpu-student/s101_api/models"
	"github.com/itpu-student/s101_api/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type seedCategory struct {
	ID    string
	Slug  string
	Emoji string
	Name  models.I18nText
	Desc  models.I18nText
}

var defaultCategories = []seedCategory{
	{
		ID:    "",
		Slug:  "all",
		Emoji: "🌍",
		Name:  models.I18nText{EN: "All", UZ: "Hammasi"},
		Desc:  models.I18nText{EN: "All", UZ: "Hammasi"},
	},
	{
		ID:    "00000000-0000-0000-0000-100000000000",
		Slug:  "restaurants",
		Emoji: "🍽️",
		Name:  models.I18nText{EN: "Restaurants", UZ: "Restoranlar"},
		Desc:  models.I18nText{EN: "Choyxona, restaurants, cafes, fast food", UZ: "Choyxona, restoranlar, kafelar, tez ovqat"},
	},
	{
		ID:    "00000000-0000-0000-0000-500000000000",
		Slug:  "sports",
		Emoji: "⚽",
		Name:  models.I18nText{EN: "Sports", UZ: "Sport"},
		Desc:  models.I18nText{EN: "Gyms, stadiums, swimming pools, golf clubs", UZ: "Zallar, stadionlar, suzish havzalari, golf klublari"},
	},
	{
		ID:    "00000000-0000-0000-0000-200000000000",
		Slug:  "auto",
		Emoji: "🚗",
		Name:  models.I18nText{EN: "Auto Services", UZ: "Avto Xizmatlar"},
		Desc:  models.I18nText{EN: "Car repair, car wash, petrol stations, car rental", UZ: "Avto ta'mir, yuvish, yoqilg'i shahobchalari, ijara"},
	},
	{
		ID:    "00000000-0000-0000-0000-400000000000",
		Slug:  "activities",
		Emoji: "🎡",
		Name:  models.I18nText{EN: "Activities", UZ: "Faoliyatlar"},
		Desc:  models.I18nText{EN: "Adventure parks, aqua parks, cinemas, amusement", UZ: "Sarguzasht parklari, aqua parklari, kinoteatrlar"},
	},
	{
		ID:    "00000000-0000-0000-0000-600000000000",
		Slug:  "tabiat",
		Emoji: "🏔️",
		Name:  models.I18nText{EN: "Nature (Tabiat)", UZ: "Tabiat"},
		Desc:  models.I18nText{EN: "National parks, botanical gardens, hiking trails", UZ: "Milliy bog'lar, botanika bog'lari, sayr yo'llari"},
	},
	{
		ID:    "00000000-0000-0000-0000-300000000000",
		Slug:  "health",
		Emoji: "🏥",
		Name:  models.I18nText{EN: "Health", UZ: "Salomatlik"},
		Desc:  models.I18nText{EN: "Clinics, hospitals, pharmacies, dental", UZ: "Klinikalar, shifoxonalar, dorixonalar, stomatologiya"},
	},
}

// SeedBootstrapAdmin creates a single admin on first boot when BootstrapAdminUsername
// and BootstrapAdminPassword are set. Skipped once any admin exists.
func SeedBootstrapAdmin(ctx context.Context) {
	username := config.Cfg.BootstrapAdminUsername
	password := config.Cfg.BootstrapAdminPassword
	if username == "" || password == "" {
		return
	}
	count, err := Admins().CountDocuments(ctx, bson.M{})
	if err != nil {
		log.Printf("admin count: %v", err)
		return
	}
	if count > 0 {
		return
	}
	hash, err := utils.HashPassword(password)
	if err != nil {
		log.Printf("bcrypt: %v", err)
		return
	}
	name := config.Cfg.BootstrapAdminName
	if name == "" {
		name = username
	}
	_, err = Admins().InsertOne(ctx, models.Admin{
		ID:           utils.NewUUID(),
		Username:     username,
		PasswordHash: hash,
		Name:         name,
		Power:        100,
		CreatedAt:    time.Now().UTC(),
	})
	if err != nil {
		log.Printf("bootstrap admin insert: %v", err)
		return
	}
	log.Printf("bootstrap admin %q created", username)
}

func SeedCategories(ctx context.Context) {
	now := time.Now().UTC()
	for _, c := range defaultCategories {
		filter := bson.M{"_id": c.ID}
		update := bson.M{
			"$setOnInsert": bson.M{
				"_id":        c.ID,
				"created_at": now,
			},
			"$set": bson.M{
				"slug":       c.Slug,
				"emoji":      c.Emoji,
				"name":       c.Name,
				"desc":       c.Desc,
				"updated_at": now,
			},
		}
		if _, err := Categories().UpdateOne(ctx, filter, update, options.Update().SetUpsert(true)); err != nil {
			log.Printf("seed category %s: %v", c.Slug, err)
		}
	}
	log.Println("categories seeded")
}
