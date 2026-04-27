package bot

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/itpu-student/s101_api/config"
	"github.com/itpu-student/s101_api/db"
	appmodels "github.com/itpu-student/s101_api/models"
	"github.com/itpu-student/s101_api/utils"
	"go.mongodb.org/mongo-driver/bson"
)

// Package-level bot instance — set by Start, read by SendDM. nil before Start
// runs, or when TG_BOT_TOKEN is empty.
var globalBot *bot.Bot

// Start launches the TG bot long-poller. It blocks until ctx is canceled.
// If TG_BOT_TOKEN is empty we simply log and return — handy for local dev.
func Start(ctx context.Context) {
	if config.Cfg.TGBotToken == "" {
		log.Println("TG_BOT_TOKEN not set — bot disabled")
		return
	}

	opts := []bot.Option{
		bot.WithDefaultHandler(defaultHandler),
	}
	b, err := bot.New(config.Cfg.TGBotToken, opts...)
	if err != nil {
		log.Printf("bot init failed: %v", err)
		return
	}
	globalBot = b

	// /start — ask the user to share their contact.
	b.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeExact, startHandler)

	log.Println("telegram bot started")
	b.Start(ctx)
}

// SendDM sends a plain-text DM to the given Telegram user. No-op if the bot
// isn't running (TG_BOT_TOKEN unset) or telegramID is empty / unparsable.
// Errors are returned but callers typically log+swallow.
func SendDM(ctx context.Context, telegramID string, text string) error {
	if globalBot == nil || telegramID == "" {
		return nil
	}
	chatID, err := strconv.ParseInt(telegramID, 10, 64)
	if err != nil {
		return err
	}
	_, err = globalBot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   text,
	})
	return err
}

func startHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}
	kb := &models.ReplyKeyboardMarkup{
		Keyboard: [][]models.KeyboardButton{
			{{Text: "Share my phone number", RequestContact: true}},
		},
		ResizeKeyboard:  true,
		OneTimeKeyboard: true,
	}
	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "Welcome to BuYelpUz! Please share your phone number to receive a login code.",
		ReplyMarkup: kb,
	})
}

// defaultHandler handles contact messages and everything else.
func defaultHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}
	msg := update.Message
	if msg.Contact == nil {
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: msg.Chat.ID,
			Text:   "Please tap the button to share your phone number.",
		})
		return
	}

	// Ignore contacts not belonging to the sender.
	if msg.Contact.UserID != msg.From.ID {
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: msg.Chat.ID,
			Text:   "Please share your own phone, not someone else's.",
		})
		return
	}

	tgID := strconv.FormatInt(msg.From.ID, 10)
	phone := msg.Contact.PhoneNumber
	var username *string
	if u := msg.From.Username; u != "" {
		username = &u
	}

	code := utils.NewOTP6()
	now := time.Now().UTC()
	otp := appmodels.OTPCode{
		ID:         utils.NewUUIDv7(),
		TelegramID: tgID,
		Phone:      phone,
		Username:   username,
		FirstName:  msg.From.FirstName,
		Code:       code,
		ExpiresAt:  now.Add(config.Cfg.OTP_TTL),
		Used:       false,
		CreatedAt:  now,
	}
	// Invalidate prior unused OTPs for this TG user.
	_, _ = db.OTPCodes().UpdateMany(ctx,
		bson.M{"telegram_id": tgID, "used": false},
		bson.M{"$set": bson.M{"used": true}},
	)
	if _, err := db.OTPCodes().InsertOne(ctx, otp); err != nil {
		log.Printf("otp insert failed: %v", err)
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: msg.Chat.ID,
			Text:   "Something went wrong, please try again.",
		})
		return
	}
	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: msg.Chat.ID,
		Text: fmt.Sprintf(
			"Your login code is: %s\n\nEnter it on buyelp.uz. It expires in %d minutes.",
			code, int(config.Cfg.OTP_TTL.Minutes()),
		),
	})
}
