# HOW_TO_CHECK.md — Telegram Bot: Phrase Checker Implementation Plan

## Overview

A Telegram bot written in **Go** that quizzes students on German↔English phrase translation pairs from the tandem worksheet. It scores answers, gives structured feedback, and tracks progress per user.

---

## Scoring Rules

| Result | Points | Condition |
|--------|--------|-----------|
| Fully correct | 1.0 | Every word matches exactly, in correct position |
| Partially correct | 0.5 | ≥50% of words are correct (exact or near-match) |
| Incorrect | 0.0 | <50% of words are correct |

### Word-Level Matching Rules

1. **Exact match** — word matches the expected word at its position (case-insensitive).
2. **Near-match (typo tolerance)** — word differs by exactly **1 character** from expected (1 insertion, deletion, or substitution — Levenshtein distance = 1). Counts as correct for scoring, but flagged as *partially correct* in feedback.
3. **Position match** — a word is checked first at its expected position. If not found there, it is searched across all positions to distinguish *wrong position* from *missing*.

---

## Feedback Categories (per answer)

After each answer, the bot reports:

- ✅ **Correct words** — matched exactly or within 1-char typo at the right position
- 🔀 **Position fails** — word exists in the answer but is at the wrong position
- ❌ **Missing words** — word from the expected phrase not found anywhere in the user's answer
- 🔤 **Partially correct words** — matched with 1-char edit distance (typo flagged)
- ➕ **Incorrect/redundant words** — words in user's answer not matching any expected word

---

## Data Model

```go
// Phrase represents a bilingual phrase pair
type Phrase struct {
    ID      int
    German  string
    English string
}

// QuizSession tracks an active student session
type QuizSession struct {
    UserID      int64
    Phrases     []Phrase   // shuffled subset for this session
    Current     int        // index into Phrases
    Direction   string     // "de->en" or "en->de"
    Score       float64
    Total       int
}

// WordResult holds per-word analysis
type WordResult struct {
    Word      string
    Expected  string
    Position  int
    Status    string // "correct", "typo", "wrong_position", "missing", "redundant"
}

// AnswerResult holds full analysis of one answer
type AnswerResult struct {
    UserAnswer    string
    ExpectedPhrase string
    Words         []WordResult
    Points        float64    // 0.0, 0.5, or 1.0
    Feedback      string     // formatted feedback string for Telegram
}
```

---

## Core Algorithm: `CheckAnswer`

```go
func CheckAnswer(expected, userAnswer string) AnswerResult {
    expectedWords := tokenize(expected)   // lowercase, strip punctuation
    userWords     := tokenize(userAnswer)

    n := len(expectedWords)
    matched := make([]bool, n)         // expected slots claimed
    usedUser := make([]bool, len(userWords))

    results := make([]WordResult, n)
    for i := range results {
        results[i] = WordResult{Expected: expectedWords[i], Position: i, Status: "missing"}
    }

    // --- Pass 1: Exact match at correct position ---
    for i := 0; i < n && i < len(userWords); i++ {
        if userWords[i] == expectedWords[i] {
            results[i].Status = "correct"
            results[i].Word   = userWords[i]
            matched[i]        = true
            usedUser[i]       = true
        }
    }

    // --- Pass 2: Near-match (Levenshtein ≤ 1) at correct position ---
    for i := 0; i < n && i < len(userWords); i++ {
        if matched[i] || usedUser[i] { continue }
        if levenshtein(userWords[i], expectedWords[i]) == 1 {
            results[i].Status = "typo"
            results[i].Word   = userWords[i]
            matched[i]        = true
            usedUser[i]       = true
        }
    }

    // --- Pass 3: Wrong position (exact or near-match elsewhere) ---
    for i := 0; i < n; i++ {
        if matched[i] { continue }
        for j := 0; j < len(userWords); j++ {
            if usedUser[j] { continue }
            dist := levenshtein(userWords[j], expectedWords[i])
            if dist == 0 || dist == 1 {
                results[i].Status = "wrong_position"
                results[i].Word   = userWords[j]
                usedUser[j]       = true
                matched[i]        = true
                break
            }
        }
    }

    // --- Collect redundant words (user words never matched) ---
    var redundant []WordResult
    for j, used := range usedUser {
        if !used {
            redundant = append(redundant, WordResult{
                Word:   userWords[j],
                Status: "redundant",
            })
        }
    }

    // --- Scoring ---
    correctCount := 0
    for _, r := range results {
        if r.Status == "correct" || r.Status == "typo" || r.Status == "wrong_position" {
            correctCount++
        }
    }

    var points float64
    ratio := float64(correctCount) / float64(n)
    switch {
    case ratio == 1.0 && countStatus(results, "typo") == 0:
        points = 1.0
    case ratio >= 0.5:
        points = 0.5
    default:
        points = 0.0
    }

    return AnswerResult{
        UserAnswer:     userAnswer,
        ExpectedPhrase: expected,
        Words:          append(results, redundant...),
        Points:         points,
        Feedback:       buildFeedback(results, redundant, expected, points),
    }
}
```

---

## Feedback Formatter: `buildFeedback`

```go
func buildFeedback(results []WordResult, redundant []WordResult, expected string, points float64) string {
    var sb strings.Builder

    // Score line
    switch points {
    case 1.0:
        sb.WriteString("✅ *Perfect! +1 point*\n\n")
    case 0.5:
        sb.WriteString("🟡 *Partially correct. +0.5 points*\n\n")
    default:
        sb.WriteString("❌ *Incorrect. 0 points*\n\n")
    }

    // Per-category feedback
    var typos, wrongPos, missing []string
    for _, r := range results {
        switch r.Status {
        case "typo":
            typos = append(typos, fmt.Sprintf("`%s` → `%s`", r.Word, r.Expected))
        case "wrong_position":
            wrongPos = append(wrongPos, fmt.Sprintf("`%s` (expected at position %d)", r.Word, r.Position+1))
        case "missing":
            missing = append(missing, fmt.Sprintf("`%s`", r.Expected))
        }
    }
    var redund []string
    for _, r := range redundant {
        redund = append(redund, fmt.Sprintf("`%s`", r.Word))
    }

    if len(typos) > 0 {
        sb.WriteString("🔤 *Typos (1 char off):*\n")
        for _, t := range typos { sb.WriteString("  • " + t + "\n") }
        sb.WriteString("\n")
    }
    if len(wrongPos) > 0 {
        sb.WriteString("🔀 *Wrong position:*\n")
        for _, w := range wrongPos { sb.WriteString("  • " + w + "\n") }
        sb.WriteString("\n")
    }
    if len(missing) > 0 {
        sb.WriteString("❌ *Missing words:*\n")
        for _, m := range missing { sb.WriteString("  • " + m + "\n") }
        sb.WriteString("\n")
    }
    if len(redund) > 0 {
        sb.WriteString("➕ *Redundant/incorrect words:*\n")
        for _, r := range redund { sb.WriteString("  • " + r + "\n") }
        sb.WriteString("\n")
    }

    sb.WriteString(fmt.Sprintf("📖 *Correct answer:*\n_%s_", expected))
    return sb.String()
}
```

---

## Levenshtein Distance (1-char check)

```go
// levenshtein returns edit distance between two strings (early-exit at 2 for performance)
func levenshtein(a, b string) int {
    ra, rb := []rune(a), []rune(b)
    la, lb := len(ra), len(rb)
    if abs(la-lb) > 1 { return 2 } // early exit — can't be ≤1

    dp := make([][]int, la+1)
    for i := range dp {
        dp[i] = make([]int, lb+1)
        dp[i][0] = i
    }
    for j := 0; j <= lb; j++ { dp[0][j] = j }

    for i := 1; i <= la; i++ {
        for j := 1; j <= lb; j++ {
            if ra[i-1] == rb[j-1] {
                dp[i][j] = dp[i-1][j-1]
            } else {
                dp[i][j] = 1 + min3(dp[i-1][j], dp[i][j-1], dp[i-1][j-1])
            }
        }
    }
    return dp[la][lb]
}
```

---

## Telegram Bot Flow

```
/start
  → Welcome message + language direction choice
     [🇩🇪→🇬🇧 DE to EN]  [🇬🇧→🇩🇪 EN to DE]  [🔀 Mixed]

/quiz
  → Starts a session (all 15 phrases, shuffled)
  → Sends phrase prompt: "Translate: *Leider müssen wir Ihnen mitteilen...*"
  → User types translation
  → Bot calls CheckAnswer(), sends formatted feedback
  → Next phrase automatically, or /skip to skip

/score
  → Shows current session: X / 15 phrases done, Y.Y / 15.0 points

/end
  → Ends session, shows final score summary table

/repeat
  → Restart with same or new direction
```

---

## Project Structure

```
phrase-bot/
├── main.go              # Bot entrypoint, Telegram handler setup
├── bot/
│   └── handlers.go      # /start, /quiz, /score, /end command handlers
├── checker/
│   ├── checker.go       # CheckAnswer(), buildFeedback(), levenshtein()
│   ├── tokenizer.go     # tokenize(): lowercase, strip punctuation/articles
│   └── checker_test.go  # Unit tests for all scoring cases
├── phrases/
│   └── phrases.go       # Hardcoded []Phrase slice from CONTENT.md
├── session/
│   └── session.go       # In-memory QuizSession store (map[int64]*QuizSession)
├── go.mod
└── go.sum
```

---

## Key Dependencies

```go
// go.mod
require (
    github.com/go-telegram-bot-api/telegram-bot-api/v5 v5.5.1
)
```

---

## Example Interaction

```
Bot: Translate this phrase:
     "Wir möchten die Bestellung stornieren."

User: We wish cancell the order

Bot: 🟡 Partially correct. +0.5 points

     🔤 Typos (1 char off):
       • `cancell` → `cancel`

     📖 Correct answer:
     _We wish to cancel the order._
```

```
Bot: Translate this phrase:
     "Die Lieferung ist durch Feuchtigkeit beschädigt worden."

User: The order was damaged by moisture during transport

Bot: ❌ Incorrect. 0 points

     🔀 Wrong position:
       • `damaged` (expected at position 4)

     ❌ Missing words:
       • `delivery`  • `been`  • `in`  • `transit`

     ➕ Redundant/incorrect words:
       • `order`  • `was`  • `during`  • `transport`

     📖 Correct answer:
     _The delivery has been damaged by moisture/dirt/rough handling in transit._
```

---

## Testing Checklist

| Test Case | Expected Points |
|-----------|----------------|
| Exact match | 1.0 |
| 1-char typo in one word | 0.5 |
| All words present, wrong order | 0.5 |
| Half words correct | 0.5 |
| Fewer than 50% words correct | 0.0 |
| Empty answer | 0.0 |
| Extra redundant words only | 0.5 (if ≥50% base words present) |
| 2-char typo | treated as incorrect word |
