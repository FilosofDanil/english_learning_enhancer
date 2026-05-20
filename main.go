package main

import (
	"log"
	"os"
	"path/filepath"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	applbot "english_learn_tg_bot/internal/bot"
	"english_learn_tg_bot/internal/checker"
	"english_learn_tg_bot/internal/config"
	"english_learn_tg_bot/internal/content"
	"english_learn_tg_bot/internal/session"
)

type telegramChecker struct{}

func (telegramChecker) Check(expected, user string) applbot.FeedbackResult {
	ar := checker.CheckAnswer(expected, user)
	return applbot.FeedbackResult{Points: ar.Points, Feedback: ar.Feedback}
}

type fileCatalog struct{ path string }

func (f fileCatalog) AllPhrases() ([]content.Phrase, error) {
	return (&content.Loader{Path: f.path}).ParseFile()
}

func (f fileCatalog) RawMarkdown() ([]byte, error) {
	return os.ReadFile(filepath.Clean(f.path))
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	api, err := tgbotapi.NewBotAPI(cfg.TelegramToken)
	if err != nil {
		log.Fatalf("telegram bot: %v", err)
	}

	lvl := cfg.LogLevel
	api.Debug = lvl == "debug"

	log.Printf("authorized as %s", api.Self.UserName)

	store := session.NewMemory()
	catalog := fileCatalog{path: cfg.ContentPath}
	h := applbot.NewHandlers(api, store, catalog, telegramChecker{}, log.Default())

	applbot.RunLongPolling(api, h, cfg.PollTimeout)
}
