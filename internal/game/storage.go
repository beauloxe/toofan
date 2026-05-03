package game

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var dataDir string

func init() {
	configDir, _ := os.UserConfigDir()
	dataDir = filepath.Join(configDir, "toofan")
}

// SaveResult appends a line to results.txt — human readable
// format: 2026-04-01 22:18 |  85 wpm | 97.5% | 30s | words:english:english_1k | 83 raw | 2 err
func SaveResult(s Stats, duration int, mode string, language string, wordSet string) {
	os.MkdirAll(dataDir, 0755)

	f, err := os.OpenFile(
		filepath.Join(dataDir, "results.txt"),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644,
	)
	if err != nil {
		return
	}
	defer f.Close()

	label := mode
	if mode == "code" {
		label = "code:" + language
	} else if mode == "words" && language != "" {
		label = "words:" + language
		if wordSet != "" {
			label += ":" + wordSet
		}
	}

	fmt.Fprintf(f, "%s | %3.0f wpm | %5.1f%% | %3ds | %s | %3.0f raw | %d err\n",
		time.Now().Format("2006-01-02 15:04"),
		s.WPM, s.Accuracy, duration, label, s.Raw, s.Mistakes,
	)
}

func GetPB(duration int, mode string) float64 {
	path := filepath.Join(dataDir, "pb.txt")
	f, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		parts := strings.SplitN(scanner.Text(), "=", 2)
		if len(parts) == 2 {
			key := parts[0]
			val, _ := strconv.ParseFloat(parts[1], 64)

			idx := strings.LastIndex(key, "-")
			if idx > 0 && key[:idx] == mode {
				dur, _ := strconv.Atoi(key[idx+1:])
				if dur == duration {
					return val
				}
			}
		}
	}
	return 0
}

func SavePB(duration int, mode string, wpm float64) {
	os.MkdirAll(dataDir, 0755)
	path := filepath.Join(dataDir, "pb.txt")

	pbs := make(map[string]float64)
	if f, err := os.Open(path); err == nil {
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			parts := strings.SplitN(scanner.Text(), "=", 2)
			if len(parts) == 2 {
				val, _ := strconv.ParseFloat(parts[1], 64)
				pbs[parts[0]] = val
			}
		}
		f.Close()
	}

	key := fmt.Sprintf("%s-%d", mode, duration)
	pbs[key] = wpm

	f, err := os.Create(path)
	if err != nil {
		return
	}
	defer f.Close()

	for k, val := range pbs {
		fmt.Fprintf(f, "%s=%.0f\n", k, val)
	}
}

func LoadConfig() (duration int, mode string, language string, wordSet string, themeName string, testMode string, wordCount int) {
	duration, mode, language, wordSet, themeName, testMode, wordCount = 30, "words", "english", "easy", "tokyonight", "time", 25

	path := filepath.Join(dataDir, "config.txt")
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		parts := strings.SplitN(scanner.Text(), "=", 2)
		if len(parts) != 2 {
			continue
		}
		switch parts[0] {
		case "duration":
			duration, _ = strconv.Atoi(parts[1])
		case "mode":
			mode = parts[1]
		case "lang":
			language = parts[1]
		case "difficulty", "wordset":
			wordSet = parts[1]
		case "theme":
			themeName = parts[1]
		case "test":
			testMode = parts[1]
		case "wordcount":
			wordCount, _ = strconv.Atoi(parts[1])
		}
	}
	return
}

func SaveConfig(duration int, mode string, language string, wordSet string, themeName string, testMode string, wordCount int) {
	os.MkdirAll(dataDir, 0755)
	f, err := os.Create(filepath.Join(dataDir, "config.txt"))
	if err != nil {
		return
	}
	defer f.Close()

	fmt.Fprintf(f, "duration=%d\nmode=%s\nlang=%s\nwordset=%s\ntheme=%s\ntest=%s\nwordcount=%d\n",
		duration, mode, language, wordSet, themeName, testMode, wordCount)
}

// SplitBundle parses a bundled backup file (sections marked with "### filename")
// and returns a map of filename -> content.
func SplitBundle(content string) map[string]string {
	sections := make(map[string]string)
	var currentName string
	var buf strings.Builder

	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, "### ") {
			if currentName != "" {
				sections[currentName] = strings.TrimRight(buf.String(), "\n")
			}
			currentName = strings.TrimPrefix(line, "### ")
			buf.Reset()
		} else if currentName != "" {
			buf.WriteString(line + "\n")
		}
	}
	if currentName != "" {
		sections[currentName] = strings.TrimRight(buf.String(), "\n")
	}
	return sections
}

func SaveBackup() (string, error) {
	backupDir := filepath.Join(dataDir, "backups")
	os.MkdirAll(backupDir, 0755)

	stamp := time.Now().Format("2006-01-02_15-04")
	dest := filepath.Join(backupDir, fmt.Sprintf("toofan_backup_%s.txt", stamp))

	var bundle strings.Builder
	for _, name := range []string{"results.txt", "pb.txt", "config.txt"} {
		data, err := os.ReadFile(filepath.Join(dataDir, name))
		if err != nil {
			continue
		}
		bundle.WriteString("### " + name + "\n")
		bundle.Write(data)
		bundle.WriteString("\n")
	}
	if err := os.WriteFile(dest, []byte(bundle.String()), 0644); err != nil {
		return "", err
	}
	return dest, nil
}

func RestoreBackup(src string) error {
	raw, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	for name, data := range SplitBundle(string(raw)) {
		os.WriteFile(filepath.Join(dataDir, name), []byte(data), 0644)
	}
	return nil
}

func ListBackups() ([]string, string) {
	backupDir := filepath.Join(dataDir, "backups")
	files, _ := filepath.Glob(filepath.Join(backupDir, "toofan_backup_*.txt"))
	for i, j := 0, len(files)-1; i < j; i, j = i+1, j-1 {
		files[i], files[j] = files[j], files[i]
	}
	return files, backupDir
}
