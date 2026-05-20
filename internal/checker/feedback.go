package checker

import (
	"fmt"
	"strings"
)

// BuildFeedbackRu formats HOW_TO_CHECK-style sections in Russian (Telegram ParseMode HTML).
func BuildFeedbackRu(results []WordResult, redundant []WordResult, expected string, points float64) string {
	var sb strings.Builder

	switch points {
	case 1.0:
		sb.WriteString("✅ <b>Идеально! +1 балл</b>\n\n")
	case 0.5:
		sb.WriteString("🟡 <b>Частично верно. +0.5 балла</b>\n\n")
	default:
		sb.WriteString("❌ <b>Неверно. 0 баллов</b>\n\n")
	}

	var typos, wrongPos, missing []string
	for _, r := range results {
		switch r.Status {
		case StatusTypo:
			typos = append(typos, fmt.Sprintf("<code>%s</code> → <code>%s</code>",
				htmlEscape(r.Word), htmlEscape(r.Expected)))
		case StatusWrongPosition:
			wrongPos = append(wrongPos, fmt.Sprintf("<code>%s</code> (ожидалось на позиции %d)",
				htmlEscape(r.Word), r.Position+1))
		case StatusMissing:
			missing = append(missing, fmt.Sprintf("<code>%s</code>", htmlEscape(r.Expected)))
		}
	}
	var redund []string
	for _, r := range redundant {
		redund = append(redund, fmt.Sprintf("<code>%s</code>", htmlEscape(r.Word)))
	}

	if len(typos) > 0 {
		sb.WriteString("🔤 <b>Опечатки (отличие в 1 символ):</b>\n")
		for _, t := range typos {
			sb.WriteString("  • " + t + "\n")
		}
		sb.WriteString("\n")
	}
	if len(wrongPos) > 0 {
		sb.WriteString("🔀 <b>Не на своём месте:</b>\n")
		for _, w := range wrongPos {
			sb.WriteString("  • " + w + "\n")
		}
		sb.WriteString("\n")
	}
	if len(missing) > 0 {
		sb.WriteString("❌ <b>Пропущенные слова:</b>\n")
		for _, m := range missing {
			sb.WriteString("  • " + m + "\n")
		}
		sb.WriteString("\n")
	}
	if len(redund) > 0 {
		sb.WriteString("➕ <b>Лишние или неверные слова:</b>\n")
		for _, r := range redund {
			sb.WriteString("  • " + r + "\n")
		}
		sb.WriteString("\n")
	}

	sb.WriteString("📖 <b>Правильный ответ:</b>\n")
	sb.WriteString("<i>")
	sb.WriteString(htmlEscape(expected))
	sb.WriteString("</i>")
	return sb.String()
}

func htmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}
