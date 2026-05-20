package content

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestParseCONTENTFile(t *testing.T) {
	_, here, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(here), "..", ".."))
	md := filepath.Join(root, "CONTENT.md")

	got, err := (&Loader{Path: md}).ParseFile()
	if err != nil {
		t.Fatalf("parse CONTENT.md: %v", err)
	}
	const wantRows = 15
	if len(got) != wantRows {
		t.Fatalf("expected %d phrases, got %d", wantRows, len(got))
	}
	if got[0].German == "" || got[0].English == "" {
		t.Fatalf("first row malformed: %+v", got[0])
	}
}
