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

func TestBackspaceWordDeletesPreviousWord(t *testing.T) {
	g := New(0, "words", "english", "easy")
	g.SetText("hello world")

	for _, ch := range "hello wor" {
		g.TypeChar(ch)
	}
	g.BackspaceWord()

	if g.Input() != "hello " {
		t.Fatalf("input = %q, want %q", g.Input(), "hello ")
	}
}

func TestBackspaceWordDeletesWordBeforeSpace(t *testing.T) {
	g := New(0, "words", "english", "easy")
	g.SetText("hello world")

	for _, ch := range "hello " {
		g.TypeChar(ch)
	}
	g.BackspaceWord()

	if g.Input() != "" {
		t.Fatalf("input = %q, want empty", g.Input())
	}
}

func TestBackspaceWordClearsLiveErrorsOnly(t *testing.T) {
	g := New(0, "words", "english", "easy")
	g.SetText("hello world")

	for _, ch := range "hello wxr" {
		g.TypeChar(ch)
	}
	g.BackspaceWord()

	if g.Input() != "hello " {
		t.Fatalf("input = %q, want %q", g.Input(), "hello ")
	}
	if len(g.Errors()) != 0 {
		t.Fatalf("errors = %#v, want none", g.Errors())
	}

	for _, ch := range "wor" {
		g.TypeChar(ch)
	}
	if len(g.ErrorWords()) != 1 || g.ErrorWords()[0] != "world" {
		t.Fatalf("ErrorWords() = %#v, want %#v", g.ErrorWords(), []string{"world"})
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
