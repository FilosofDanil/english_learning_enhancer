package checker

type WordStatus string

const (
	StatusCorrect       WordStatus = "correct"
	StatusTypo          WordStatus = "typo"
	StatusWrongPosition WordStatus = "wrong_position"
	StatusMissing       WordStatus = "missing"
	StatusRedundant     WordStatus = "redundant"
)

// WordResult describes one expected-slot or surplus user token.
type WordResult struct {
	Word     string
	Expected string
	Position int
	Status   WordStatus
}

// AnswerResult aggregates scoring and Russian feedback HTML.
type AnswerResult struct {
	UserAnswer     string
	ExpectedPhrase string
	Words          []WordResult
	Points         float64
	Feedback       string // HTML-safe chunk (ParseMode HTML)
}

// CheckAnswer analyzes userAnswer against expected reference English phrase (HOW_TO_CHECK algorithm).
func CheckAnswer(expected, userAnswer string) AnswerResult {
	expectedWords := Tokenize(expected)
	userWords := Tokenize(userAnswer)

	n := len(expectedWords)
	if n == 0 {
		// Degenerate content; CONTENT.md always yields tokens.
		return AnswerResult{
			UserAnswer:     userAnswer,
			ExpectedPhrase: expected,
			Points:         0,
			Feedback:       BuildFeedbackRu(nil, nil, expected, 0),
		}
	}

	matched := make([]bool, n)
	usedUser := make([]bool, len(userWords))

	results := make([]WordResult, n)
	for i := range results {
		results[i] = WordResult{Expected: expectedWords[i], Position: i, Status: StatusMissing}
	}

	// Pass 1: exact at position
	for i := 0; i < n && i < len(userWords); i++ {
		if userWords[i] == expectedWords[i] {
			results[i].Status = StatusCorrect
			results[i].Word = userWords[i]
			matched[i] = true
			usedUser[i] = true
		}
	}

	// Pass 2: Levenshtein == 1 at position
	for i := 0; i < n && i < len(userWords); i++ {
		if matched[i] || usedUser[i] {
			continue
		}
		if Levenshtein(userWords[i], expectedWords[i]) == 1 {
			results[i].Status = StatusTypo
			results[i].Word = userWords[i]
			matched[i] = true
			usedUser[i] = true
		}
	}

	// Pass 3: wrong position (exact or 1-char elsewhere)
	for i := 0; i < n; i++ {
		if matched[i] {
			continue
		}
		for j := 0; j < len(userWords); j++ {
			if usedUser[j] {
				continue
			}
			d := Levenshtein(userWords[j], expectedWords[i])
			if d == 0 || d == 1 {
				results[i].Status = StatusWrongPosition
				results[i].Word = userWords[j]
				usedUser[j] = true
				matched[i] = true
				break
			}
		}
	}

	var redundant []WordResult
	for j, used := range usedUser {
		if !used {
			redundant = append(redundant, WordResult{
				Word:   userWords[j],
				Status: StatusRedundant,
			})
		}
	}

	correctCount := 0
	for _, r := range results {
		if r.Status == StatusCorrect || r.Status == StatusTypo || r.Status == StatusWrongPosition {
			correctCount++
		}
	}

	ratio := float64(correctCount) / float64(n)
	var points float64
	switch {
	case ratio == 1 && countTypoStatuses(results) == 0:
		points = 1.0
	case ratio >= 0.5:
		points = 0.5
	default:
		points = 0.0
	}

	feedback := BuildFeedbackRu(results, redundant, expected, points)
	return AnswerResult{
		UserAnswer:     userAnswer,
		ExpectedPhrase: expected,
		Words:          append(results, redundant...),
		Points:         points,
		Feedback:       feedback,
	}
}

func countTypoStatuses(results []WordResult) int {
	c := 0
	for _, r := range results {
		if r.Status == StatusTypo {
			c++
		}
	}
	return c
}
