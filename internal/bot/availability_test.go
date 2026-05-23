package bot

import (
	"net/http"
	"testing"
	"time"
)

func TestTelegramServerAvailability(t *testing.T) {
	client := &http.Client{Timeout: 5 * time.Second}

	// Using an intentionally invalid bot token endpoint:
	// any non-5xx response means Telegram API is reachable.
	resp, err := client.Get("https://api.telegram.org/botinvalid-token/getMe")
	if err != nil {
		t.Fatalf("telegram API is unreachable: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusInternalServerError {
		t.Fatalf("telegram API is reachable but unhealthy, got status %d", resp.StatusCode)
	}
}
