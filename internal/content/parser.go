package content

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Phrase is a bilingual pair from CONTENT.md (German prompts, English answers).
type Phrase struct {
	ID      int
	German  string
	English string
}

// Loader reads phrase pairs from a markdown table file (CONTENT.md shape).
type Loader struct {
	Path string
}

// ParseFile reads CONTENT.md-like table rows into phrases.
func (l Loader) ParseFile() ([]Phrase, error) {
	data, err := os.ReadFile(filepath.Clean(l.Path))
	if err != nil {
		return nil, fmt.Errorf("read content %q: %w", l.Path, err)
	}
	return ParseMarkdownTable(string(data))
}

// ParseMarkdownTable extracts phrases from CONTENT.md markdown table body.
func ParseMarkdownTable(markdown string) ([]Phrase, error) {
	const header = "| German | English |"
	const sepMarker = "---"

	lines := strings.Split(markdown, "\n")
	idx := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == header {
			idx = i
			break
		}
	}
	if idx < 0 || idx+1 >= len(lines) {
		return nil, fmt.Errorf("CONTENT.md: header %q not found", header)
	}

	separator := strings.TrimSpace(lines[idx+1])
	if !strings.HasPrefix(separator, "|") || !strings.Contains(separator, sepMarker) {
		return nil, fmt.Errorf("CONTENT.md: invalid table separator after header")
	}

	var out []Phrase
	id := 1
	for _, line := range lines[idx+2:] {
		row := strings.TrimSpace(line)
		if row == "" || row[0] != '|' {
			continue
		}
		inner := strings.TrimSpace(row[1:])
		inner = strings.TrimSuffix(inner, "|")
		inner = strings.TrimSpace(inner)
		fields := strings.Split(inner, "|")
		for i := range fields {
			fields[i] = strings.TrimSpace(fields[i])
		}
		if len(fields) < 2 {
			continue
		}
		de, en := fields[0], fields[1]
		if de == "" || en == "" {
			continue
		}

		out = append(out, Phrase{
			ID:      id,
			German:  de,
			English: en,
		})
		id++
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("CONTENT.md: no phrase rows parsed")
	}
	return out, nil
}
