package bot

import (
	"fmt"
	"log"
	"math/rand"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"english_learn_tg_bot/internal/content"
	"english_learn_tg_bot/internal/i18n"
	"english_learn_tg_bot/internal/session"
)

// AnswerChecker evaluates user English answer against canonical English.
type AnswerChecker interface {
	Check(expectedEnglish, user string) FeedbackResult
}

// FeedbackResult is what telegram layer needs after check.
type FeedbackResult struct {
	Points   float64
	Feedback string // HTML-safe for ParseMode HTML
}

// PhraseCatalog loads phrases from CONTENT.md and raw file contents for /print.
type PhraseCatalog interface {
	AllPhrases() ([]content.Phrase, error)
	RawMarkdown() ([]byte, error)
}

// Handlers aggregates telegram transport and domain ports.
type Handlers struct {
	API     *tgbotapi.BotAPI
	Session session.Store
	Catalog PhraseCatalog
	Check   AnswerChecker
	Log     *log.Logger
}

// NewHandlers builds a handler group with sane defaults on logger.
func NewHandlers(api *tgbotapi.BotAPI, store session.Store, cat PhraseCatalog, chk AnswerChecker, lg *log.Logger) *Handlers {
	if lg == nil {
		lg = log.Default()
	}
	return &Handlers{
		API:     api,
		Session: store,
		Catalog: cat,
		Check:   chk,
		Log:     lg,
	}
}

// Dispatch routes one Telegram update to command switch or plain text answers.
func (h *Handlers) Dispatch(u *tgbotapi.Update) {
	if u == nil || u.Message == nil {
		return
	}
	msg := u.Message
	if msg.From == nil {
		return
	}
	userID := msg.From.ID
	chatID := msg.Chat.ID

	if msg.IsCommand() {
		switch msg.Command() {
		case "start":
			h.start(chatID, userID)
		case "test":
			h.startTest(chatID, userID)
		case "print":
			h.printAll(chatID)
		case "stop":
			h.stop(chatID, userID)
		default:
			h.unknown(chatID)
		}
		return
	}

	if strings.TrimSpace(msg.Text) != "" {
		h.answerText(chatID, userID, msg.Text)
	}
}

func (h *Handlers) start(chatID, userID int64) {
	_ = userID // reserved if per-user greetings are needed later
	h.replyHTML(chatID, i18n.StartWelcome)
}

func (h *Handlers) startTest(chatID, userID int64) {
	if _, ok := h.Session.Get(userID); ok {
		h.replyHTML(chatID, i18n.MsgTestRestarted)
	}

	phrases, err := h.Catalog.AllPhrases()
	if err != nil || len(phrases) == 0 {
		if h.Log != nil {
			h.Log.Printf("load phrases: %v", err)
		}
		h.replyHTML(chatID, i18n.ErrLoadPhrases)
		return
	}

	shuffled := shufflePhrases(phrases)
	s := h.Session.Start(userID, shuffled)
	sendCurrentPrompt(h, chatID, s)
}

func shufflePhrases(in []content.Phrase) []content.Phrase {
	out := make([]content.Phrase, len(in))
	copy(out, in)
	rand.Shuffle(len(out), func(i, j int) { out[i], out[j] = out[j], out[i] })
	return out
}

func sendCurrentPrompt(h *Handlers, chatID int64, s *session.QuizSession) {
	if s.Current >= len(s.Phrases) {
		finalizeQuiz(h, chatID, s)
		return
	}
	p := s.Phrases[s.Current]
	promptDE := `<blockquote>` + htmlEscape(p.German) + `</blockquote>`
	text := fmt.Sprintf("%s\n\n%s",
		htmlEscape(i18n.MsgTranslatePrompt),
		promptDE,
	)
	h.replyHTML(chatID, text)
}

func finalizeQuiz(h *Handlers, chatID int64, s *session.QuizSession) {
	total := float64(len(s.Phrases))
	text := fmt.Sprintf(i18n.MsgSessionCompleteFmt, s.Score, total, len(s.Phrases))
	h.replyHTML(chatID, text)
	h.Session.Remove(s.UserID)
}

func progressLineRU(s *session.QuizSession) string {
	done := s.Current
	total := len(s.Phrases)
	maxEarnedPossible := float64(done)
	return fmt.Sprintf(i18n.MsgProgLine, done, total, s.Score, maxEarnedPossible)
}

func (h *Handlers) answerText(chatID, userID int64, txt string) {
	s, ok := h.Session.Get(userID)
	if !ok || s == nil || s.Current >= len(s.Phrases) {
		h.replyHTML(chatID, i18n.MsgNoActiveUseTest)
		return
	}

	p := s.Phrases[s.Current]
	fr := h.Check.Check(p.English, txt)

	s.Score += fr.Points
	s.Current++

	line := strings.TrimSuffix(fr.Feedback+"\n\n"+progressLineRU(s)+"\n", "\n")

	if s.Current >= len(s.Phrases) {
		h.replyHTML(chatID, line)
		finalizeQuiz(h, chatID, s)
		return
	}
	h.replyHTML(chatID, line)
	sendCurrentPrompt(h, chatID, s)
}
func (h *Handlers) stop(chatID, userID int64) {
	s, ok := h.Session.Get(userID)
	if !ok || s == nil {
		h.replyHTML(chatID, i18n.MsgNoActiveUseTest)
		return
	}

	done := s.Current
	total := len(s.Phrases)
	maxSoFar := float64(done)

	var text string
	if done == 0 {
		text = "<b>" + i18n.MsgFinalStopped + "</b>\n\nВы ещё не ответили ни на один вопрос. Используй /test когда будешь готов."
	} else {
		text = fmt.Sprintf(i18n.MsgStopSummaryFmt, s.Score, maxSoFar, done, total)
	}
	h.replyHTML(chatID, text)
	h.Session.Remove(userID)
}

func (h *Handlers) unknown(chatID int64) {
	h.replyHTML(chatID, i18n.MsgUnknownCmd)
}

func (h *Handlers) printAll(chatID int64) {
	b, err := h.Catalog.RawMarkdown()
	if err != nil {
		if h.Log != nil {
			h.Log.Printf("print raw markdown: %v", err)
		}
		h.replyHTML(chatID, i18n.ErrLoadPhrases)
		return
	}

	text := string(b)
	cfg := tgbotapi.MessageConfig{
		BaseChat:              tgbotapi.BaseChat{ChatID: chatID},
		DisableWebPagePreview: true,
		Text:                  "",
		ParseMode:             "",
	}
	chunks := splitTelegramChunks(text, 3900)

	for _, ch := range chunks {
		cfg.Text = ch
		if _, err := h.API.Send(cfg); err != nil && h.Log != nil {
			h.Log.Printf("send print chunk: %v", err)
		}
	}
}

func htmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

func (h *Handlers) replyHTML(chatID int64, text string) {
	cfg := tgbotapi.MessageConfig{
		BaseChat:              tgbotapi.BaseChat{ChatID: chatID},
		DisableWebPagePreview: true,
		Text:                  "",
		ParseMode:             tgbotapi.ModeHTML,
	}
	for _, chunk := range splitTelegramChunks(text, 3900) {
		cfg.Text = chunk
		if _, err := h.API.Send(cfg); err != nil && h.Log != nil {
			h.Log.Printf("send html chunk: %v", err)
		}
	}
}

func splitTelegramChunks(s string, max int) []string {
	if max <= 0 {
		return []string{s}
	}
	if len(s) <= max {
		return []string{s}
	}
	runes := []rune(s)
	var out []string
	i := 0
	for len(runes)-i > max {
		out = append(out, string(runes[i:i+max]))
		i += max
	}
	out = append(out, string(runes[i:]))
	return out
}
