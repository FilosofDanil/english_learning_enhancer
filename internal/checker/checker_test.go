package checker

import (
	"testing"
)

func TestExactPhraseOnePoint(t *testing.T) {
	expected := "one two three four"
	got := CheckAnswer(expected, expected)
	if got.Points != 1.0 {
		t.Fatalf("want 1.0 got %.1f (%s)", got.Points, got.Feedback)
	}
}

func TestPartialHalfWords(t *testing.T) {
	expected := "one two three four"
	got := CheckAnswer(expected, "one two wrong wrong")
	if got.Points != 0.5 {
		t.Fatalf("want 0.5 got %.1f", got.Points)
	}
}

func TestIncorrectBelowHalf(t *testing.T) {
	expected := "one two three four"
	got := CheckAnswer(expected, "foo bar baz extra")
	if got.Points != 0.0 {
		t.Fatalf("want 0.0 got %.1f", got.Points)
	}
}

func TestEmptyAnswerZero(t *testing.T) {
	got := CheckAnswer("one two three four", "")
	if got.Points != 0 {
		t.Fatalf("want 0 got %.2f", got.Points)
	}
}

func TestSingleTypoMeansHalfUnlessPerfect(t *testing.T) {
	expected := "onlyword"
	got := CheckAnswer(expected, "onlywort")
	if got.Points != 0.5 {
		t.Fatalf("want 0.5 got %.2f (%s)", got.Points, got.Feedback)
	}
}

func TestTwoSubstitutionsMeansZero(t *testing.T) {
	// Needs Levenshtein distance 2 versus the lone expected token "foo".
	got := CheckAnswer("foo", "faa")
	if got.Points != 0.0 {
		t.Fatalf("want 0.0 got %.2f", got.Points)
	}
}

func TestExtraRedundantKeepsEnoughWordsForHalf(t *testing.T) {
	expected := "alfa beta gamma delta"
	got := CheckAnswer(expected, "alfa beta zzzz zzzz")
	if got.Points != 0.5 {
		t.Fatalf("want 0.5 got %.2f", got.Points)
	}
}

func TestTokenizeSplitsSlashes(t *testing.T) {
	toks := Tokenize("moisture/dirt,in-transit!")
	if len(toks) < 4 {
		t.Fatalf("want >=4 tokens, got %v", toks)
	}
}

func TestMissingWordsAreMarkedAndScored(t *testing.T) {
	expected := "one two three four"
	got := CheckAnswer(expected, "one two")

	if got.Points != 0.5 {
		t.Fatalf("want 0.5 for 2/4 matched words, got %.2f", got.Points)
	}
	if len(got.Words) < 4 {
		t.Fatalf("want at least 4 word results, got %d", len(got.Words))
	}
	if got.Words[2].Status != StatusMissing || got.Words[3].Status != StatusMissing {
		t.Fatalf("expected missing statuses at positions 2 and 3, got %s and %s", got.Words[2].Status, got.Words[3].Status)
	}
}

func TestAllWordsCorrectWithPunctuationAndCase(t *testing.T) {
	expected := "One two three four"
	got := CheckAnswer(expected, "one, TWO; three... four!")

	if got.Points != 1.0 {
		t.Fatalf("want 1.0 for fully correct normalized answer, got %.2f (%s)", got.Points, got.Feedback)
	}
	for i, w := range got.Words {
		if w.Status != StatusCorrect {
			t.Fatalf("word at index %d should be correct, got %s", i, w.Status)
		}
	}
}
