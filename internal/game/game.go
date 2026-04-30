package game

import (
	"strings"
	"time"
	"unicode/utf8"

	"github.com/vyrx-dev/toofan/internal/lang"
)

type Stats struct {
	WPM      float64
	Raw      float64
	Accuracy float64
	Chars    int
	Mistakes int // total wrong keystrokes, including corrected ones
}

type Game struct {
	text       string
	Snippet    lang.Snippet
	input      string
	errors     map[int]bool // unfixed errors — used for live display coloring
	mistakeAt  map[int]bool // every wrong keystroke ever — never cleared by backspace
	started    bool
	duration   int
	CodeMode   bool // true = snippet-based typing, false = standard words
	difficulty string
	wordSet    string

	elapsed   time.Duration
	lastTyped time.Time
	LastTick  time.Time
}

// Accessors for fields that TUI needs to read
func (g *Game) Text() string           { return g.text }
func (g *Game) Input() string          { return g.input }
func (g *Game) Errors() map[int]bool   { return g.errors }
func (g *Game) Started() bool          { return g.started }
func (g *Game) Duration() int          { return g.duration }
func (g *Game) Elapsed() time.Duration { return g.elapsed }
func (g *Game) SetText(s string)       { g.text = normalizeTabs(s) }
func (g *Game) WordSet() string        { return g.wordSet }

func New(duration int, mode string, language string, difficulty string) *Game {
	g := &Game{
		duration:   duration, // 0 means infinite mode (tied to length of snippet)
		errors:     make(map[int]bool),
		mistakeAt:  make(map[int]bool),
		difficulty: difficulty,
	}

	if mode == "code" {
		if !lang.HasSnippets(language) {
			language = lang.DefaultCodeName()
		}
		g.CodeMode = true
		g.Snippet = lang.RandomSnippet(language, difficulty)
		g.text = normalizeTabs(g.Snippet.Content)
	} else {
		if !lang.HasWords(language) {
			language = lang.DefaultWordName()
		}
		words, wordSet := lang.RandomWordsWithSet(language, difficulty, 200)
		g.wordSet = wordSet
		g.text = strings.Join(words, " ")
	}

	return g
}

func (g *Game) Tick(now time.Time) {
	if !g.started || g.Finished() {
		return
	}

	dt := now.Sub(g.LastTick)
	g.LastTick = now
	g.elapsed += dt
}

// normalizeTabs replaces tabs with spaces so position tracking and
// rendering stay perfectly aligned in the terminal.
func normalizeTabs(s string) string {
	return strings.ReplaceAll(s, "\t", "    ")
}

func isStartOfLine(s string) bool {
	for i := len(s); i > 0; {
		r, size := utf8.DecodeLastRuneInString(s[:i])
		i -= size
		if r == '\n' {
			return true
		}
		if r != ' ' {
			return false
		}
	}
	return true
}

func runeLen(s string) int {
	return utf8.RuneCountInString(s)
}

func trimLastRune(s string) string {
	if s == "" {
		return s
	}
	_, size := utf8.DecodeLastRuneInString(s)
	return s[:len(s)-size]
}

func lastRune(s string) (rune, bool) {
	if s == "" {
		return 0, false
	}
	r, _ := utf8.DecodeLastRuneInString(s)
	return r, true
}

func (g *Game) TypeChar(ch rune) {
	if g.Finished() {
		return
	}

	if !g.started {
		g.started = true
		g.LastTick = time.Now()
	}

	g.lastTyped = time.Now()

	textRunes := []rune(g.text)
	pos := runeLen(g.input)
	if pos >= len(textRunes) {
		return
	}

	if g.CodeMode {
		if textRunes[pos] == '\n' {
			if ch == ' ' {
				ch = '\n' // allow space to act as enter at the end of a line
			}
			if ch != '\n' {
				return // strict newline enforcement
			}
		} else if ch == '\n' {
			return
		}
	}

	g.input += string(ch)
	if textRunes[pos] != ch {
		g.errors[pos] = true
		g.mistakeAt[pos] = true
	}

	// auto-skip newlines and leading indentation
	for runeLen(g.input) < len(textRunes) {
		next := textRunes[runeLen(g.input)]
		if next == '\n' {
			if g.CodeMode {
				break
			}
			g.input += "\n"
		} else if next == ' ' && isStartOfLine(g.input) {
			g.input += " "
		} else {
			break
		}
	}
}

func (g *Game) Backspace() {
	if g.input == "" {
		return
	}

	g.lastTyped = time.Now()

	// skip back over auto-inserted newlines and leading whitespace
	for g.input != "" {
		last, ok := lastRune(g.input)
		if !ok {
			break
		}
		if last == '\n' {
			if g.CodeMode {
				break
			}
			g.input = trimLastRune(g.input)
		} else if last == ' ' && isStartOfLine(trimLastRune(g.input)) {
			g.input = trimLastRune(g.input)
		} else {
			break
		}
	}

	if g.input != "" {
		pos := runeLen(g.input) - 1
		g.input = trimLastRune(g.input)
		delete(g.errors, pos)
		// mistakeAt is NOT cleared — tracks lifetime errors
	}
}

func (g *Game) TimeLeft() int {
	if !g.started {
		return g.duration
	}
	if g.duration == 0 {
		return int(g.elapsed.Seconds()) // Count up in infinite mode
	}
	left := g.duration - int(g.elapsed.Seconds())
	if left < 0 {
		return 0
	}
	return left
}

func (g *Game) Finished() bool {
	if !g.started {
		return false
	}
	if g.duration == 0 {
		// Infinite duration finishes when the text is completed
		return runeLen(g.input) >= runeLen(g.text)
	}
	return g.TimeLeft() == 0 || runeLen(g.input) >= runeLen(g.text)
}

func (g *Game) Stats() Stats {
	if !g.started {
		return Stats{}
	}

	mins := g.elapsed.Minutes()
	if mins < 0.001 {
		mins = 0.001
	}

	// count only real typed chars (skip auto-inserted newlines)
	total := 0
	unfixed := 0
	mistakes := 0
	for i, ch := range []rune(g.input) {
		if ch != '\n' {
			total++
			if g.errors[i] {
				unfixed++
			}
			if g.mistakeAt[i] {
				mistakes++
			}
		}
	}

	// WPM uses unfixed errors (corrections already penalized via time spent)
	raw := float64(total) / 5.0 / mins
	wpm := float64(total-unfixed) / 5.0 / mins
	if wpm < 0 {
		wpm = 0
	}

	// Accuracy uses ALL mistakes — correcting doesn't erase the miss
	acc := 100.0
	if total > 0 {
		acc = float64(total-mistakes) / float64(total) * 100
		if acc < 0 {
			acc = 0
		}
	}

	return Stats{WPM: wpm, Raw: raw, Accuracy: acc, Chars: total, Mistakes: mistakes}
}

// ErrorWords returns the words where the user made mistakes (word mode only).
// Includes corrected mistakes so users can see where they struggled.
func (g *Game) ErrorWords() []string {
	if g.CodeMode {
		return nil
	}
	words := strings.Split(g.text, " ")
	pos := 0
	inputLen := runeLen(g.input)
	var result []string
	seen := make(map[string]bool)
	for _, word := range words {
		if pos >= inputLen {
			break
		}
		hasErr := false
		for j := 0; j < runeLen(word) && pos+j < inputLen; j++ {
			if g.mistakeAt[pos+j] {
				hasErr = true
				break
			}
		}
		if hasErr && !seen[word] {
			result = append(result, word)
			seen[word] = true
		}
		pos += runeLen(word) + 1 // +1 for space separator
	}
	return result
}

func (g *Game) Reset(mode string, language string, difficulty string) {
	g.difficulty = difficulty
	if mode == "code" {
		if !lang.HasSnippets(language) {
			language = lang.DefaultCodeName()
		}
		g.CodeMode = true
		g.wordSet = ""
		g.Snippet = lang.RandomSnippet(language, difficulty)
		g.text = normalizeTabs(g.Snippet.Content)
	} else {
		if !lang.HasWords(language) {
			language = lang.DefaultWordName()
		}
		g.CodeMode = false
		words, wordSet := lang.RandomWordsWithSet(language, difficulty, 200)
		g.wordSet = wordSet
		g.text = strings.Join(words, " ")
	}
	g.input = ""
	g.errors = make(map[int]bool)
	g.mistakeAt = make(map[int]bool)
	g.started = false
	g.elapsed = 0
	g.lastTyped = time.Time{}
	g.LastTick = time.Time{}
}
