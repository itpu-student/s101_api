package config

import (
	"log"
	"os"
	"time"

	"github.com/go-shafaq/timep"
	"github.com/joho/godotenv"
)

type Config struct {
	Port string

	MongoURI string
	MongoDB  string

	JWTSecret string
	JWTTTL    time.Duration

	TGBotToken string

	OTP_TTL        time.Duration
	SOATOLookupURL string
	Env            string

	BootstrapAdminUsername string
	BootstrapAdminPassword string
	BootstrapAdminName     string
}

var Cfg *Config

func Load() *Config {
	_ = godotenv.Load()

	Cfg = &Config{
		Port:           getEnv("PORT", "8080"),
		MongoURI:       getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDB:        getEnv("MONGO_DB", "buyelp"),
		JWTSecret:      getEnv("JWT_SECRET", "dev-secret-change-me"),
		JWTTTL:         getEnvDur("JWT_TTL", "720h"),
		TGBotToken:     getEnv("TG_BOT_TOKEN", ""),
		OTP_TTL:        getEnvDur("OTP_TTL", "5m"),
		SOATOLookupURL: getEnv("SOATO_LOOKUP_URL", ""),
		Env:            getEnv("ENV", "development"),

		BootstrapAdminUsername: getEnv("BOOTSTRAP_ADMIN_USERNAME", "admin"),
		BootstrapAdminPassword: getEnv("BOOTSTRAP_ADMIN_PASSWORD", "admin"),
		BootstrapAdminName:     getEnv("BOOTSTRAP_ADMIN_NAME", "Admin"),
	}

	if Cfg.JWTSecret == "dev-secret-change-me" && Cfg.Env == "production" {
		log.Fatal("JWT_SECRET must be set in production")
	}
	return Cfg
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getEnvDur(key, fallback string) time.Duration {
	v, _ := os.LookupEnv(key)
	if v == "" {
		v = fallback
	}

	dur, err := timep.ParseDuration(v)
	if err != nil {
		log.Printf("invalid duration for %s: %v, using fallback", key, err)
		panic(err)
	}

	return dur
}
