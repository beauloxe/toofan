package tui

import "testing"

func TestVisibleLineWindowKeepsNextLineAvailable(t *testing.T) {
	tests := []struct {
		name      string
		lineCount int
		curLine   int
		visible   int
		wantTop   int
		wantBot   int
	}{
		{name: "first line", lineCount: 5, curLine: 0, visible: 3, wantTop: 0, wantBot: 3},
		{name: "middle line", lineCount: 5, curLine: 1, visible: 3, wantTop: 0, wantBot: 3},
		{name: "last visible line slides", lineCount: 5, curLine: 2, visible: 3, wantTop: 1, wantBot: 4},
		{name: "continues sliding", lineCount: 5, curLine: 3, visible: 3, wantTop: 2, wantBot: 5},
		{name: "clamps at end", lineCount: 5, curLine: 4, visible: 3, wantTop: 2, wantBot: 5},
		{name: "all lines fit", lineCount: 2, curLine: 1, visible: 3, wantTop: 0, wantBot: 2},
		{name: "empty", lineCount: 0, curLine: 0, visible: 3, wantTop: 0, wantBot: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTop, gotBot := visibleLineWindow(tt.lineCount, tt.curLine, tt.visible)
			if gotTop != tt.wantTop || gotBot != tt.wantBot {
				t.Fatalf("visibleLineWindow() = (%d, %d), want (%d, %d)", gotTop, gotBot, tt.wantTop, tt.wantBot)
			}
		})
	}
}
