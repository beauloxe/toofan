package tui

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/vyrx-dev/toofan/internal/theme"
)

type profileData struct {
	Tests         int
	Time          time.Duration
	Best          map[string]map[int]float64 // mode -> dur -> wpm
	Recent        []testEntry
	Activity      map[string]int
	RecentAvg     float64
	RecentCodeAvg float64
}

type testEntry struct {
	Date   time.Time
	WPM    float64
	Dur    int
	Acc    float64
	Mode   string
	Raw    float64
	Errors int
}

func loadProfile() profileData {
	pd := profileData{
		Best:     make(map[string]map[int]float64),
		Activity: make(map[string]int),
	}
	pd.Best["words"] = make(map[int]float64)
	pd.Best["code"] = make(map[int]float64)

	configDir, err := os.UserConfigDir()
	if err != nil {
		return pd
	}
	dataDir := filepath.Join(configDir, "toofan")

	f, err := os.Open(filepath.Join(dataDir, "results.txt"))
	if err != nil {
		return pd
	}
	defer f.Close()

	var all []testEntry
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		e, ok := parseResultLine(sc.Text())
		if !ok {
			continue
		}
		all = append(all, e)
		pd.Tests++
		pd.Time += time.Duration(e.Dur) * time.Second

		mode, _, _ := splitResultMode(e.Mode)
		if pd.Best[mode] == nil {
			pd.Best[mode] = make(map[int]float64)
		}
		if e.WPM > pd.Best[mode][e.Dur] {
			pd.Best[mode][e.Dur] = e.WPM
		}
		pd.Activity[e.Date.Format("2006-01-02")]++
	}

	if len(all) > 80 {
		pd.Recent = all[len(all)-80:]
	} else {
		pd.Recent = all
	}

	var wordsTests []testEntry
	var codeTests []testEntry
	for _, e := range all {
		mode, _, _ := splitResultMode(e.Mode)
		if mode == "words" {
			wordsTests = append(wordsTests, e)
		} else {
			codeTests = append(codeTests, e)
		}
	}

	pd.RecentAvg = avgWPM(wordsTests)
	pd.RecentCodeAvg = avgWPM(codeTests)

	return pd
}

// avgWPM returns the average WPM of the last 10 tests (or fewer if < 10 exist).
func avgWPM(tests []testEntry) float64 {
	if len(tests) == 0 {
		return 0
	}
	n := 10
	if len(tests) < n {
		n = len(tests)
	}
	sum := 0.0
	for i := len(tests) - n; i < len(tests); i++ {
		sum += tests[i].WPM
	}
	return sum / float64(n)
}

func parseResultLine(line string) (testEntry, bool) {
	parts := strings.Split(line, "|")
	if len(parts) < 5 {
		return testEntry{}, false
	}

	date, err := time.Parse("2006-01-02 15:04", strings.TrimSpace(parts[0]))
	if err != nil {
		return testEntry{}, false
	}

	wpmStr := strings.TrimSpace(parts[1])
	wpmStr = strings.TrimSuffix(wpmStr, "wpm")
	wpmStr = strings.TrimSpace(wpmStr)
	wpm, _ := strconv.ParseFloat(wpmStr, 64)

	accStr := strings.TrimSpace(parts[2])
	accStr = strings.TrimSuffix(accStr, "%")
	accStr = strings.TrimSpace(accStr)
	acc, _ := strconv.ParseFloat(accStr, 64)

	durStr := strings.TrimSpace(parts[3])
	durStr = strings.TrimSuffix(durStr, "s")
	durStr = strings.TrimSpace(durStr)
	dur, _ := strconv.Atoi(durStr)

	modeStr := strings.TrimSpace(parts[4])

	var raw float64
	var errors int
	if len(parts) >= 6 {
		rawStr := strings.TrimSpace(parts[5])
		rawStr = strings.TrimSuffix(rawStr, "raw")
		rawStr = strings.TrimSpace(rawStr)
		raw, _ = strconv.ParseFloat(rawStr, 64)
	}
	if len(parts) >= 7 {
		errStr := strings.TrimSpace(parts[6])
		errStr = strings.TrimSuffix(errStr, "err")
		errStr = strings.TrimSpace(errStr)
		errors, _ = strconv.Atoi(errStr)
	}

	return testEntry{Date: date, WPM: wpm, Dur: dur, Acc: acc, Mode: modeStr, Raw: raw, Errors: errors}, true
}

func splitResultMode(mode string) (string, string, string) {
	parts := strings.Split(mode, ":")
	if len(parts) >= 2 {
		wordSet := ""
		if len(parts) >= 3 {
			wordSet = parts[2]
		}
		return parts[0], parts[1], wordSet
	}
	if mode == "code" {
		return "code", "", ""
	}
	return "words", "english", ""
}

func (m model) handleProfile(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.active = screenTyping
	return m, nil
}

func truncateLang(lang string) string {
	if len(lang) > 9 {
		return lang[:9]
	}
	return lang
}

func rank(wpm float64) string {
	switch {
	case wpm >= 120:
		return "toofan"
	case wpm >= 80:
		return "tryhard"
	case wpm >= 50:
		return "mid"
	case wpm >= 30:
		return "noob"
	default:
		return "grandma"
	}
}

func (m model) viewProfile(p theme.Palette) string {
	dim := lipgloss.NewStyle().Foreground(p.Foreground)
	val := lipgloss.NewStyle().Foreground(p.Typed).Bold(true)
	hi := lipgloss.NewStyle().Foreground(p.Accent)

	title := val.Render("_toofan")

	paneStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(p.Foreground).
		Padding(1, 2)
	horizontalPadding := 3

	hours := int(m.prof.Time.Hours())
	mins := int(m.prof.Time.Minutes()) % 60
	timeVal := fmt.Sprintf("%d", mins)
	timeUnit := "m"
	if hours > 0 {
		timeVal = fmt.Sprintf("%dh %d", hours, mins)
	}

	wordsAvgVal := "-"
	wordsAvgUnit := ""
	if m.prof.RecentAvg > 0 {
		wordsAvgVal = fmt.Sprintf("%.0f", m.prof.RecentAvg)
		wordsAvgUnit = "wpm"
	}

	codeAvgVal := "-"
	codeAvgUnit := ""
	if m.prof.RecentCodeAvg > 0 {
		codeAvgVal = fmt.Sprintf("%.0f", m.prof.RecentCodeAvg)
		codeAvgUnit = "wpm"
	}

	accVal := "-"
	accUnit := ""
	if len(m.prof.Recent) > 0 {
		var totalAcc float64
		for _, e := range m.prof.Recent {
			totalAcc += e.Acc
		}
		accVal = fmt.Sprintf("%.0f", totalAcc/float64(len(m.prof.Recent)))
		accUnit = "%"
	}

	formatRow := func(k, v, u string) string {
		formattedVal := v + u

		keyBlock := lipgloss.NewStyle().Width(13).Align(lipgloss.Left).Render(dim.Render(k))
		valBlock := lipgloss.NewStyle().Width(8).Align(lipgloss.Left).Render(val.Render(formattedVal))

		return lipgloss.JoinHorizontal(lipgloss.Left, keyBlock, valBlock)
	}

	overview := lipgloss.JoinVertical(lipgloss.Left,
		hi.Render("overview"),
		"",
		formatRow("tests", fmt.Sprintf("%d", m.prof.Tests), ""),
		formatRow("time", timeVal, timeUnit),
		formatRow("words avg", wordsAvgVal, wordsAvgUnit),
		formatRow("code avg", codeAvgVal, codeAvgUnit),
		formatRow("accuracy", accVal, accUnit),
	)

	durStyle := lipgloss.NewStyle().Width(6).Align(lipgloss.Left)
	colStyle := lipgloss.NewStyle().Width(8).Align(lipgloss.Center)

	headerLabels := lipgloss.JoinHorizontal(lipgloss.Left,
		durStyle.Render(""),
		colStyle.Render(dim.Render("words")),
		colStyle.Render(dim.Render("code")),
	)

	bestRowVert := func(dur string, d int) string {
		wStr := dim.Render("-")
		if w, ok := m.prof.Best["words"][d]; ok {
			wStr = val.Render(fmt.Sprintf("%.0f", w))
		}
		cStr := dim.Render("-")
		if c, ok := m.prof.Best["code"][d]; ok {
			cStr = val.Render(fmt.Sprintf("%.0f", c))
		}

		return lipgloss.JoinHorizontal(lipgloss.Left,
			durStyle.Render(dim.Render(dur)),
			colStyle.Render(wStr),
			colStyle.Render(cStr),
		)
	}

	bests := lipgloss.JoinVertical(lipgloss.Left,
		hi.Render("personal bests"),
		"",
		headerLabels,
		bestRowVert("15s", 15),
		bestRowVert("30s", 30),
		bestRowVert("60s", 60),
		bestRowVert("120s", 120),
	)

	cur := rank(m.prof.RecentAvg)
	type tier struct {
		name  string
		label string
	}
	tiers := []tier{
		{"grandma", "0-30"},
		{"noob", "30-50"},
		{"mid", "50-80"},
		{"tryhard", "80-120"},
		{"toofan", "120+"},
	}

	rankNameWidth := 0
	for _, t := range tiers {
		rankNameWidth = max(rankNameWidth, lipgloss.Width(t.name))
	}

	var rankLines []string
	for _, t := range tiers {
		paddedName := fmt.Sprintf("%-*s", rankNameWidth+5, t.name)
		paddedLabel := fmt.Sprintf("%-6s", t.label)

		var prefix, label string
		if t.name == cur {
			prefix = hi.Render("● ") + val.Render(paddedName)
			label = val.Render(paddedLabel)
		} else {
			prefix = dim.Render("● " + paddedName)
			label = dim.Render(paddedLabel)
		}

		rankLines = append(rankLines, lipgloss.JoinHorizontal(lipgloss.Left, prefix, label))
	}

	ranks := lipgloss.JoinVertical(lipgloss.Left,
		hi.Render("ranks"),
		"",
		strings.Join(rankLines, "\n"),
	)

	minPaneWidths := []int{
		lipgloss.Width(overview) + horizontalPadding,
		lipgloss.Width(bests) + horizontalPadding,
		lipgloss.Width(ranks) + horizontalPadding,
	}
	borderWidth := lipgloss.Width(paneStyle.Width(1).Render("")) - 1
	if borderWidth < 0 {
		borderWidth = 0
	}
	histData, histWidths := m.recentTestRows()

	var histRows []string
	header := tableHeader(
		tableCell{histWidths[0], "wpm", hi},
		tableCell{histWidths[1], "raw", hi},
		tableCell{histWidths[2], "accuracy", hi},
		tableCell{histWidths[3], "typos", hi},
		tableCell{histWidths[4], "mode", hi},
		tableCell{histWidths[5], "language", hi},
		tableCell{histWidths[6], "set", hi},
		tableCell{histWidths[7], "time", hi},
		tableCell{histWidths[8], "date", hi},
	)
	histRows = append(histRows, header, "")

	for _, hist := range histData {
		row := tableRow(
			tableCell{histWidths[0], hist[0], val},
			tableCell{histWidths[1], hist[1], dim},
			tableCell{histWidths[2], hist[2], dim},
			tableCell{histWidths[3], hist[3], dim},
			tableCell{histWidths[4], hist[4], dim},
			tableCell{histWidths[5], hist[5], dim},
			tableCell{histWidths[6], hist[6], dim},
			tableCell{histWidths[7], hist[7], dim},
			tableCell{histWidths[8], hist[8], dim},
		)
		histRows = append(histRows, row)
	}

	histTable := lipgloss.JoinVertical(lipgloss.Left, histRows...)
	paneWidth, wideBoxWidth := profileWidths(minPaneWidths, borderWidth, lipgloss.Width(histTable)+horizontalPadding)

	// Render all boxes first to measure actual heights
	overviewBox := paneStyle.Width(paneWidth).Render(overview)
	bestBox := paneStyle.Width(paneWidth).Render(bests)
	ranksBox := paneStyle.Width(paneWidth).Render(ranks)

	// Match heights
	maxH := lipgloss.Height(overviewBox)
	if h := lipgloss.Height(bestBox); h > maxH {
		maxH = h
	}
	if h := lipgloss.Height(ranksBox); h > maxH {
		maxH = h
	}

	overviewBox = paneStyle.Width(paneWidth).Height(maxH - 2).Render(overview)
	bestBox = paneStyle.Width(paneWidth).Height(maxH - 2).Render(bests)
	ranksBox = paneStyle.Width(paneWidth).Height(maxH - 2).Render(ranks)

	topRow := lipgloss.JoinHorizontal(lipgloss.Top, overviewBox, bestBox, ranksBox)

	histBox := paneStyle.Width(wideBoxWidth - 6).Render(
		lipgloss.JoinVertical(lipgloss.Left,
			hi.Render("recent tests"),
			"",
			histTable,
		),
	)

	heatmapStr := heatGrid(m.prof.Activity, p, wideBoxWidth-6)
	heatBox := paneStyle.Width(wideBoxWidth - 6).Render(
		lipgloss.JoinVertical(lipgloss.Left,
			hi.Render("activity map"),
			"",
			heatmapStr,
		),
	)

	body := lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		topRow,
		histBox,
		heatBox,
	)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, body)
}

func (m model) recentTestRows() ([][]string, []int) {
	headers := []string{"wpm", "raw", "accuracy", "typos", "mode", "language", "set", "time", "date"}
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = lipgloss.Width(h) + 2
	}

	limit := 10
	if len(m.prof.Recent) < limit {
		limit = len(m.prof.Recent)
	}

	var rows [][]string
	for i := len(m.prof.Recent) - 1; i >= len(m.prof.Recent)-limit; i-- {
		e := m.prof.Recent[i]
		modeType, modeLang, wordSet := splitResultMode(e.Mode)
		modeLang = truncateLang(modeLang)
		if wordSet == "" {
			wordSet = "-"
		}

		durStr := "∞"
		if e.Dur > 0 {
			durStr = fmt.Sprintf("%ds", e.Dur)
		}

		row := []string{
			fmt.Sprintf("%.0f", e.WPM),
			fmt.Sprintf("%.0f", e.Raw),
			fmt.Sprintf("%.0f%%", e.Acc),
			fmt.Sprintf("%d", e.Errors),
			modeType,
			modeLang,
			wordSet,
			durStr,
			e.Date.Format("02 Jan 15:04"),
		}
		for j, cell := range row {
			widths[j] = max(widths[j], lipgloss.Width(cell)+1)
		}
		rows = append(rows, row)
	}

	return rows, widths
}

func tableWidth(widths []int) int {
	total := 0
	for _, w := range widths {
		total += w
	}
	return total
}

func profileWidths(minPaneWidths []int, borderWidth int, histBoxWidth int) (int, int) {
	minPaneWidth := 0
	for _, w := range minPaneWidths {
		minPaneWidth = max(minPaneWidth, w)
	}

	paneWidth := minPaneWidth
	if histBoxWidth+borderWidth > (paneWidth+borderWidth)*3 {
		paneWidth = (histBoxWidth - borderWidth*2 + 2) / 3
	}
	if paneWidth < minPaneWidth {
		paneWidth = minPaneWidth
	}
	wideBoxWidth := (paneWidth+borderWidth)*3 - borderWidth
	return paneWidth, wideBoxWidth
}

func heatGrid(activity map[string]int, p theme.Palette, width int) string {
	now := time.Now()
	days := []string{"mon", "tue", "wed", "thu", "fri", "sat", "sun"}

	c0 := lipgloss.NewStyle().Foreground(lipgloss.Color("#333333"))
	c1 := lipgloss.NewStyle().Foreground(p.Accent)
	dim := lipgloss.NewStyle().Foreground(p.Foreground)

	weeks := (width - 11) / 2
	if weeks < 1 {
		weeks = 1
	}

	// Find the most recent Monday (start of current week).
	// All columns are anchored to this Monday so each column is a real calendar week.
	weekStart := now
	for weekStart.Weekday() != time.Monday {
		weekStart = weekStart.AddDate(0, 0, -1)
	}
	weekStart = time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, weekStart.Location())

	var rows []string
	for dayOffset := 0; dayOffset < 7; dayOffset++ {
		var row strings.Builder
		row.WriteString(dim.Render(fmt.Sprintf("%3s  ", days[dayOffset])))
		for w := weeks - 1; w >= 0; w-- {
			// Column w: the Monday that is w weeks before weekStart
			colMonday := weekStart.AddDate(0, 0, -w*7)
			// The actual date for this row (dayOffset 0=Mon … 6=Sun)
			d := colMonday.AddDate(0, 0, dayOffset)

			// Don't mark future dates as active
			if d.After(now) {
				row.WriteString(c0.Render("■") + " ")
				continue
			}

			if activity[d.Format("2006-01-02")] > 0 {
				row.WriteString(c1.Render("■") + " ")
			} else {
				row.WriteString(c0.Render("■") + " ")
			}
		}
		rows = append(rows, row.String())
	}
	return strings.Join(rows, "\n")
}
