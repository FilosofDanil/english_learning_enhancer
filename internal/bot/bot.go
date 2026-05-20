package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// RunLongPolling blocks forwarding Telegram updates to handlers until interrupted.
func RunLongPolling(api *tgbotapi.BotAPI, h *Handlers, timeout int) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = timeout
	updates := api.GetUpdatesChan(u)

	for upd := range updates {
		local := upd
		h.Dispatch(&local)
	}
}
