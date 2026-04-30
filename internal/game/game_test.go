package game

import (
	"testing"
	"unicode/utf8"
)

func TestTypeCharHandlesCyrillic(t *testing.T) {
	g := New(0, "words", "english", "easy")
	g.SetText("привіт світ")

	for _, ch := range g.Text() {
		g.TypeChar(ch)
	}

	if g.Input() != g.Text() {
		t.Fatalf("input = %q, want %q", g.Input(), g.Text())
	}
	if len(g.Errors()) != 0 {
		t.Fatalf("errors = %#v, want none", g.Errors())
	}
	if !g.Finished() {
		t.Fatal("game did not finish after typing full Cyrillic text")
	}

	stats := g.Stats()
	wantChars := utf8.RuneCountInString(g.Text())
	if stats.Chars != wantChars {
		t.Fatalf("chars = %d, want %d", stats.Chars, wantChars)
	}
}

func TestBackspaceHandlesCyrillic(t *testing.T) {
	g := New(0, "words", "english", "easy")
	g.SetText("пр")

	g.TypeChar('п')
	g.TypeChar('x')
	g.Backspace()

	if g.Input() != "п" {
		t.Fatalf("input = %q, want %q", g.Input(), "п")
	}
	if len(g.Errors()) != 0 {
		t.Fatalf("errors = %#v, want none", g.Errors())
	}
}

func TestErrorWordsHandlesCyrillic(t *testing.T) {
	g := New(0, "words", "english", "easy")
	g.SetText("кіт пес")

	for _, ch := range "кxт пе" {
		g.TypeChar(ch)
	}

	words := g.ErrorWords()
	if len(words) != 1 || words[0] != "кіт" {
		t.Fatalf("ErrorWords() = %#v, want %#v", words, []string{"кіт"})
	}
}
