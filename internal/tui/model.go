package tui

import (
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/vyrx-dev/toofan/internal/game"
	"github.com/vyrx-dev/toofan/internal/lang"
	"github.com/vyrx-dev/toofan/internal/theme"
)

var durations = []int{0, 15, 30, 60, 120}
var wordCounts = []int{10, 25, 50, 100}

type screen int

const (
	screenTyping screen = iota
	screenResults
	screenProfile
)

type model struct {
	active     screen
	game       *game.Game
	duration   int
	testMode   string // "time" or "words"
	wordCount  int
	mode       string // "words" or "code"
	lang       string
	difficulty string

	width, height int

	pickingDur        bool
	durCur            int
	pickingLang       bool
	langCur           int
	pickingLesson     bool
	lessonCur         int
	pickingTheme      bool
	themeCur          int
	pickingDifficulty bool
	diffCur           int
	showHelp          bool

	result        game.Stats
	pb            float64
	gotNewPB      bool
	finishedAt    time.Time
	showingErrors bool

	prof profileData

	message string
	msgTime time.Time

	pickingRestore bool
	backups        []string
	restoreCur     int
}

func New() model {
	duration, mode, language, difficulty, th, testMode, wordCount := game.LoadConfig()
	theme.Current = theme.ByName(th)
	language = languageForMode(mode, language)
	difficulty = wordSetForLanguage(mode, language, difficulty)
	testMode = testModeForMode(mode, testMode)
	wordCount = validWordCount(wordCount)

	return model{
		game:       game.NewWithWordTarget(durationForTest(testMode, duration), mode, language, difficulty, wordTargetForTest(testMode, mode, wordCount)),
		duration:   duration,
		testMode:   testMode,
		wordCount:  wordCount,
		mode:       mode,
		lang:       language,
		difficulty: difficulty,
	}
}

type tick time.Time

func (m model) isPaused() bool {
	return m.pickingDur || m.pickingLang || m.pickingLesson || m.pickingTheme || m.pickingDifficulty || m.showHelp
}

func (m model) Init() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return tick(t)
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tick:
		if m.active == screenTyping && m.game.Started() {
			if m.isPaused() {
				m.game.LastTick = time.Time(msg)
			} else {
				m.game.Tick(time.Time(msg))
				if m.game.Finished() {
					m.result = m.game.Stats()
					m.pb = game.GetPB(m.pbTarget(), m.pbMode())
					m.gotNewPB = m.result.WPM > m.pb

					durToSave := int(m.game.Elapsed().Seconds())
					if m.testMode == "time" && m.duration > 0 {
						durToSave = m.duration
					}
					game.SaveResult(m.result, durToSave, m.mode, m.lang, m.game.WordSet())
					if m.gotNewPB {
						game.SavePB(m.pbTarget(), m.pbMode(), m.result.WPM)
					}

					m.active = screenResults
					m.finishedAt = time.Now()
				}
			}
		}
		return m, tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
			return tick(t)
		})

	case tea.KeyMsg:
		if m.message != "" {
			m.message = ""
		}

		if m.pickingRestore {
			switch msg.String() {
			case "up", "k":
				if m.restoreCur > 0 {
					m.restoreCur--
				}
			case "down", "j":
				if m.restoreCur < len(m.backups)-1 {
					m.restoreCur++
				}
			case "enter":
				src := m.backups[m.restoreCur]
				if err := game.RestoreBackup(src); err == nil {
					m.message = "imported " + filepath.Base(src)
					m.msgTime = time.Now()
					m.prof = loadProfile()
				}
				m.pickingRestore = false
			case "esc", "q":
				m.pickingRestore = false
			}
			return m, nil
		}

		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "ctrl+s":
			if dest, err := game.SaveBackup(); err == nil {
				m.message = "backup saved → " + dest
				m.msgTime = time.Now()
			}
			return m, nil
		case "ctrl+r":
			files, backupDir := game.ListBackups()
			if len(files) == 0 {
				m.message = "no backups found in " + backupDir
				m.msgTime = time.Now()
				return m, nil
			}
			m.backups = files
			m.pickingRestore = true
			m.restoreCur = 0
			return m, nil
		}
		if m.showHelp {
			m.showHelp = false
			return m, nil
		}
		if m.pickingDur {
			values := durations
			if m.testMode == "words" && m.mode == "words" {
				values = wordCounts
			}
			switch msg.String() {
			case "up", "k", "left", "h":
				if m.durCur > 0 {
					m.durCur--
				}
				return m, nil
			case "down", "j", "right", "l":
				if m.durCur < len(values)-1 {
					m.durCur++
				}
				return m, nil
			case "enter":
				if m.testMode == "words" && m.mode == "words" {
					m.wordCount = wordCounts[m.durCur]
				} else {
					m.duration = durations[m.durCur]
				}
				m.pickingDur = false
				m.game = m.newGame()
				m.save()
				return m, nil
			case "esc":
				m.pickingDur = false
				return m, nil
			default:
				m.pickingDur = false
				// fallthrough to handleTyping so the key is typed
			}
		}
		if m.pickingDifficulty {
			sets := m.wordSetNames()
			if len(sets) == 0 {
				m.pickingDifficulty = false
				return m, nil
			}
			switch msg.String() {
			case "up", "k", "left", "h":
				if m.diffCur > 0 {
					m.diffCur--
				}
				return m, nil
			case "down", "j", "right", "l":
				if m.diffCur < len(sets)-1 {
					m.diffCur++
				}
				return m, nil
			case "enter":
				m.difficulty = sets[m.diffCur]
				m.pickingDifficulty = false
				m.game = m.newGame()
				m.save()
				return m, nil
			case "esc":
				m.pickingDifficulty = false
				return m, nil
			default:
				m.pickingDifficulty = false
			}
		}
		if m.pickingLang {
			return m.handlePicker(msg)
		}
		if m.pickingLesson {
			return m.handleLessonPicker(msg)
		}
		if m.pickingTheme {
			return m.handleThemePicker(msg)
		}

		switch m.active {
		case screenTyping:
			return m.handleTyping(msg)
		case screenResults:
			return m.handleResults(msg)
		case screenProfile:
			return m.handleProfile(msg)
		}
	}

	return m, nil
}

func (m model) View() string {
	if m.width == 0 {
		return ""
	}

	p := theme.Current
	var body string

	if m.pickingRestore {
		var names []string
		for _, f := range m.backups {
			names = append(names, filepath.Base(f))
		}
		body = renderList(p, "restore backup", names, nil, m.restoreCur)
	} else {
		switch m.active {
		case screenTyping:
			body = m.viewTyping(p)
		case screenProfile:
			body = m.viewProfile(p)
		case screenResults:
			body = m.viewResults(p)
		}
	}

	if m.message != "" && time.Since(m.msgTime) < 5*time.Second {
		msgStyle := lipgloss.NewStyle().Foreground(p.Background).Background(p.Success).Padding(0, 2)
		body = lipgloss.JoinVertical(lipgloss.Center,
			body,
			"", "",
			msgStyle.Render(m.message),
		)
	}

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, body)
}

func (m model) save() {
	game.SaveConfig(m.duration, m.mode, m.lang, m.difficulty, theme.Current.Name, m.testMode, m.wordCount)
}

func languageForMode(mode string, language string) string {
	if mode == "code" {
		if lang.HasSnippets(language) {
			return language
		}
		return lang.DefaultCodeName()
	}
	if lang.HasWords(language) {
		return language
	}
	return lang.DefaultWordName()
}

func wordSetForLanguage(mode string, language string, wordSet string) string {
	if mode != "words" {
		return wordSet
	}
	for _, set := range lang.WordSetNames(language) {
		if set == wordSet {
			return wordSet
		}
	}
	return lang.DefaultWordSet(language)
}

func nextDur(cur int) int {
	for i, d := range durations {
		if d == cur {
			return durations[(i+1)%len(durations)]
		}
	}
	return 30
}

func nextWordCount(cur int) int {
	for i, count := range wordCounts {
		if count == cur {
			return wordCounts[(i+1)%len(wordCounts)]
		}
	}
	return 25
}

func (m model) newGame() *game.Game {
	return game.NewWithWordTarget(durationForTest(m.testMode, m.duration), m.mode, m.lang, m.difficulty, wordTargetForTest(m.testMode, m.mode, m.wordCount))
}

func (m model) pbMode() string {
	if m.mode == "words" && m.testMode == "words" {
		return "words-count"
	}
	return m.mode
}

func (m model) pbTarget() int {
	if m.mode == "words" && m.testMode == "words" {
		return m.wordCount
	}
	return m.duration
}

func durationForTest(testMode string, duration int) int {
	if testMode == "words" {
		return 0
	}
	return duration
}

func wordTargetForTest(testMode string, mode string, wordCount int) int {
	if mode == "words" && testMode == "words" {
		return validWordCount(wordCount)
	}
	return 0
}

func testModeForMode(mode string, testMode string) string {
	if mode != "words" || testMode != "words" {
		return "time"
	}
	return "words"
}

func validWordCount(wordCount int) int {
	for _, count := range wordCounts {
		if count == wordCount {
			return wordCount
		}
	}
	return 25
}
