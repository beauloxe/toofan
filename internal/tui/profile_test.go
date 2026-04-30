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
