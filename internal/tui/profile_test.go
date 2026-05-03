package tui

import "testing"

func TestSplitResultMode(t *testing.T) {
	tests := []struct {
		name     string
		mode     string
		wantType string
		wantLang string
		wantSet  string
	}{
		{name: "legacy words", mode: "words", wantType: "words", wantLang: "english"},
		{name: "word language", mode: "words:spanish", wantType: "words", wantLang: "spanish"},
		{name: "word set", mode: "words:spanish:easy", wantType: "words", wantLang: "spanish", wantSet: "easy"},
		{name: "word count set", mode: "words25:spanish:easy", wantType: "words25", wantLang: "spanish", wantSet: "easy"},
		{name: "code language", mode: "code:go", wantType: "code", wantLang: "go"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotType, gotLang, gotSet := splitResultMode(tt.mode)
			if gotType != tt.wantType || gotLang != tt.wantLang || gotSet != tt.wantSet {
				t.Fatalf("splitResultMode(%q) = %q, %q, %q; want %q, %q, %q",
					tt.mode, gotType, gotLang, gotSet, tt.wantType, tt.wantLang, tt.wantSet)
			}
		})
	}
}

func TestWordCountFromMode(t *testing.T) {
	tests := map[string]int{
		"words":    0,
		"words10":  10,
		"words25":  25,
		"words100": 100,
		"code":     0,
	}

	for mode, want := range tests {
		if got := wordCountFromMode(mode); got != want {
			t.Fatalf("wordCountFromMode(%q) = %d; want %d", mode, got, want)
		}
	}
}
