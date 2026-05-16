package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"github.com/itpu-student/s101_api/config"
	"github.com/itpu-student/s101_api/db"
	"github.com/itpu-student/s101_api/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/api/option"
)

type AISummary struct {
	En string `json:"en"`
	Uz string `json:"uz"`
}

const (
	minReviewsForSummary = 2
	summaryMaxReviews    = 20
	summaryCacheTTL      = 24 * time.Hour
	// summaryCacheTTL     = 24 * time.Microsecond
)

func GetOrRefreshSummary(ctx context.Context, placeID string) (*AISummary, error) {
	var p models.Place
	if err := db.Places().FindOne(ctx, bson.M{"_id": placeID}).Decode(&p); err != nil {
		return nil, err
	}

	if p.AISummaryUpdatedAt != nil && time.Since(*p.AISummaryUpdatedAt) < summaryCacheTTL {
		if p.AISummary.EN != "" {
			return &AISummary{En: p.AISummary.EN, Uz: p.AISummary.UZ}, nil
		}
	}

	cur, err := db.Reviews().Find(ctx,
		bson.M{"place_id": placeID, "latest": true},
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetLimit(summaryMaxReviews),
	)
	if err != nil {
		return nil, err
	}
	var reviews []models.Review
	if err = cur.All(ctx, &reviews); err != nil {
		return nil, err
	}

	if len(reviews) < minReviewsForSummary {
		return nil, nil
	}

	summary, err := callGemini(ctx, p.Name, reviews)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	_, _ = db.Places().UpdateByID(ctx, placeID, bson.M{"$set": bson.M{
		"ai_summary.en":         summary.En,
		"ai_summary.uz":         summary.Uz,
		"ai_summary_updated_at": now,
	}})

	return summary, nil
}

func callGemini(ctx context.Context, placeName string, reviews []models.Review) (*AISummary, error) {
	apiKey := config.Cfg.GeminiAPIKey
	if apiKey == "" {
		return nil, errors.New("GEMINI_API_KEY not set")
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("gemini client: %w", err)
	}
	defer client.Close()

	var sb strings.Builder
	for _, r := range reviews {
		sb.WriteString(fmt.Sprintf("- %d stars: %s\n", r.StarRating, r.Text))
	}

	prompt := fmt.Sprintf(`You are summarizing user reviews for "%s", a local business in Uzbekistan.
Summarize the reviews below in 2-3 sentences. Be balanced, mention both positives and negatives if present.
Return ONLY valid JSON in exactly this format with no extra text:
{"en":"<English summary>","uz":"<Uzbek summary>"}

Reviews:
%s`, placeName, sb.String())

	model := client.GenerativeModel("gemini-flash-latest")
	model.GenerationConfig.ResponseMIMEType = "application/json"

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("gemini generate: %w", err)
	}

	if len(resp.Candidates) == 0 {
		return nil, errors.New("gemini returned no candidates")
	}

	cand := resp.Candidates[0]
	if len(cand.Content.Parts) == 0 {
		return nil, fmt.Errorf("gemini candidate has no parts (finish_reason=%v)", cand.FinishReason)
	}

	// Parts[0] is genai.Text — must type-assert, not fmt.Sprintf
	part, ok := cand.Content.Parts[0].(genai.Text)
	if !ok {
		return nil, fmt.Errorf("unexpected gemini part type: %T", cand.Content.Parts[0])
	}

	raw := string(part)

	var result AISummary
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return nil, fmt.Errorf("parse gemini response: %w", err)
	}

	return &result, nil
}
