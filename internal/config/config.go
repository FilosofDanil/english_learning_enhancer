package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds runtime settings loaded from the environment (.env optional).
type Config struct {
	TelegramToken string
	ContentPath   string
	PollTimeout   int
	LogLevel      string
}

// Load reads .env (if present) and environment variables into Config.
func Load() (Config, error) {
	_ = godotenv.Load()

	cfg := Config{
		ContentPath:   getEnv("CONTENT_PATH", "CONTENT.md"),
		PollTimeout:   getEnvInt("POLL_TIMEOUT", 60),
		LogLevel:      strings.ToLower(getEnv("LOG_LEVEL", "info")),
		TelegramToken: strings.TrimSpace(os.Getenv("TELEGRAM_BOT_TOKEN")),
	}
	if cfg.TelegramToken == "" {
		return Config{}, fmt.Errorf("TELEGRAM_BOT_TOKEN is required")
	}
	return cfg, nil
}

func getEnv(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	s := strings.TrimSpace(os.Getenv(key))
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 {
		return def
	}
	return n
}
