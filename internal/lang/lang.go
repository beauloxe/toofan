package lang

import (
	"embed"
	"encoding/json"
	"io/fs"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"
)

//go:embed data
var dataFS embed.FS

type Snippet struct {
	Topic   string
	Content string
}

type wordSet struct {
	Name  string
	Words []string
}

type langData struct {
	Name        string
	Words       []string
	EasyWords   []string
	MediumWords []string
	HardWords   []string
	WordSets    []wordSet
	Snippets    []Snippet
}

var languages = map[string]*langData{}

// Names holds code language names, sorted. Kept for older call sites.
var Names []string
var CodeNames []string
var WordNames []string

type monkeytypeWords struct {
	Name  string   `json:"name"`
	Words []string `json:"words"`
}

func parseWords(path string, raw []byte) ([]string, string) {
	if !utf8.Valid(raw) {
		return nil, ""
	}

	label := cleanSetLabel(strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)))

	if strings.EqualFold(filepath.Ext(path), ".json") {
		var mt monkeytypeWords
		if err := json.Unmarshal(raw, &mt); err == nil && len(mt.Words) > 0 {
			if strings.TrimSpace(mt.Name) != "" {
				label = cleanSetLabel(mt.Name)
			}
			return cleanWords(mt.Words), label
		}

		var words []string
		if err := json.Unmarshal(raw, &words); err == nil && len(words) > 0 {
			return cleanWords(words), label
		}
		return nil, ""
	}

	return cleanWords(strings.Fields(string(raw))), label
}

func cleanWords(words []string) []string {
	out := make([]string, 0, len(words))
	seen := make(map[string]bool)
	for _, word := range words {
		for _, field := range strings.Fields(word) {
			if !utf8.ValidString(field) || seen[field] {
				continue
			}
			out = append(out, field)
			seen[field] = true
		}
	}
	return out
}

func cleanSetLabel(label string) string {
	label = strings.TrimSpace(label)
	label = strings.ReplaceAll(label, "|", "-")
	label = strings.ReplaceAll(label, ":", "-")
	return label
}

func loadWordData(ld *langData, path string, raw []byte, bucket string) {
	words, label := parseWords(path, raw)
	if len(words) == 0 {
		return
	}

	if label == "" {
		label = bucket
	}
	if label == "" {
		label = "words"
	}
	addWordSet(ld, label, words)

	switch bucket {
	case "easy":
		ld.EasyWords = append(ld.EasyWords, words...)
	case "medium":
		ld.MediumWords = append(ld.MediumWords, words...)
	case "hard":
		ld.HardWords = append(ld.HardWords, words...)
	}
	ld.Words = append(ld.Words, words...)
}

func loadEmbeddedWordFile(ld *langData, path string, bucket string) {
	raw, err := fs.ReadFile(dataFS, path)
	if err != nil {
		return
	}
	loadWordData(ld, path, raw, bucket)
}

func loadRuntimeWordFile(ld *langData, path string, bucket string) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return
	}
	loadWordData(ld, path, raw, bucket)
}

func addWordSet(ld *langData, label string, words []string) {
	for i := range ld.WordSets {
		if ld.WordSets[i].Name == label {
			ld.WordSets[i].Words = append(ld.WordSets[i].Words, words...)
			return
		}
	}
	ld.WordSets = append(ld.WordSets, wordSet{Name: label, Words: words})
}

func wordBucket(filename string) (string, bool) {
	ext := filepath.Ext(filename)
	base := strings.TrimSuffix(filename, ext)

	switch base {
	case "easy", "medium", "hard":
		return base, true
	case "words":
		return "", true
	}

	if strings.EqualFold(ext, ".json") {
		return "", true
	}
	return "", false
}

// parseLesson extracts a snippet from a lesson file.
// Leading comments are stripped from the typed content.
// Supports // (go/js/dart), # (shell), and -- (lua) comment styles.
// The first "Topic:" comment becomes the display heading.
func parseLesson(content string) []Snippet {
	lines := strings.Split(strings.TrimSpace(content), "\n")
	var topic string
	codeStart := 0

	// scan leading comments for metadata, find where code begins
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// check if this is a comment line (any supported prefix)
		isComment := strings.HasPrefix(trimmed, "//") ||
			strings.HasPrefix(trimmed, "#") ||
			strings.HasPrefix(trimmed, "--")

		if !isComment {
			break
		}

		// extract "Topic:" from any comment style
		if topic == "" {
			for _, prefix := range []string{"// Topic: ", "# Topic: ", "-- Topic: "} {
				if strings.HasPrefix(trimmed, prefix) {
					topic = strings.TrimSpace(strings.TrimPrefix(trimmed, prefix))
				}
			}
		}

		codeStart = i + 1
	}

	// skip blank lines between comments and code
	for codeStart < len(lines) && strings.TrimSpace(lines[codeStart]) == "" {
		codeStart++
	}

	codeText := strings.TrimSpace(strings.Join(lines[codeStart:], "\n"))
	if len(codeText) == 0 {
		return nil
	}
	if topic == "" {
		topic = "Code Snippet"
	}
	return []Snippet{{Topic: topic, Content: codeText}}
}

func ensureLang(name string) *langData {
	ld, ok := languages[name]
	if !ok {
		ld = &langData{Name: name}
		languages[name] = ld
	}
	return ld
}

func appendUnique(list []string, name string) []string {
	for _, item := range list {
		if item == name {
			return list
		}
	}
	return append(list, name)
}

func loadEmbeddedHumanLanguages() {
	entries, err := fs.ReadDir(dataFS, "data/human")
	if err != nil {
		return
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		ld := ensureLang(e.Name())

		files, _ := fs.ReadDir(dataFS, "data/human/"+e.Name())
		for _, entry := range files {
			if entry.IsDir() {
				continue
			}
			if bucket, ok := wordBucket(entry.Name()); ok {
				path := "data/human/" + e.Name() + "/" + entry.Name()
				loadEmbeddedWordFile(ld, path, bucket)
			}
		}

		if len(ld.Words) > 0 {
			sortWordSets(ld.WordSets)
			WordNames = appendUnique(WordNames, e.Name())
		}
	}
}

func loadEmbeddedProgrammingLanguages() {
	entries, err := fs.ReadDir(dataFS, "data/programming")
	if err != nil {
		return
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		ld := ensureLang(e.Name())

		files, _ := fs.ReadDir(dataFS, "data/programming/"+e.Name())
		for _, entry := range files {
			if entry.IsDir() {
				continue
			}

			path := "data/programming/" + e.Name() + "/" + entry.Name()
			if raw, err := fs.ReadFile(dataFS, path); err == nil {
				snips := parseLesson(string(raw))
				ld.Snippets = append(ld.Snippets, snips...)
			}
		}

		if len(ld.Snippets) > 0 {
			CodeNames = appendUnique(CodeNames, e.Name())
		}
	}
}

func runtimeRoot() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(configDir, "toofan", "lang")
}

func loadRuntimeHumanLanguages(root string) {
	entries, err := os.ReadDir(filepath.Join(root, "human"))
	if err != nil {
		return
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		ld := ensureLang(e.Name())

		dir := filepath.Join(root, "human", e.Name())
		files, _ := os.ReadDir(dir)
		for _, entry := range files {
			if entry.IsDir() {
				continue
			}
			if bucket, ok := wordBucket(entry.Name()); ok {
				loadRuntimeWordFile(ld, filepath.Join(dir, entry.Name()), bucket)
			}
		}

		if len(ld.Words) > 0 {
			sortWordSets(ld.WordSets)
			WordNames = appendUnique(WordNames, e.Name())
		}
	}
}

func loadRuntimeProgrammingLanguages(root string) {
	entries, err := os.ReadDir(filepath.Join(root, "programming"))
	if err != nil {
		return
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		ld := ensureLang(e.Name())

		dir := filepath.Join(root, "programming", e.Name())
		files, _ := os.ReadDir(dir)
		for _, entry := range files {
			if entry.IsDir() {
				continue
			}
			raw, err := os.ReadFile(filepath.Join(dir, entry.Name()))
			if err != nil {
				continue
			}
			snips := parseLesson(string(raw))
			ld.Snippets = append(ld.Snippets, snips...)
		}

		if len(ld.Snippets) > 0 {
			CodeNames = appendUnique(CodeNames, e.Name())
		}
	}
}

func init() {
	loadEmbedded()
	if root := runtimeRoot(); root != "" {
		loadRuntime(root)
	}
	finalizeIndexes()
}

func loadEmbedded() {
	loadEmbeddedHumanLanguages()
	loadEmbeddedProgrammingLanguages()
}

func loadRuntime(root string) {
	loadRuntimeHumanLanguages(root)
	loadRuntimeProgrammingLanguages(root)
}

func finalizeIndexes() {
	sort.Strings(CodeNames)
	sort.Strings(WordNames)
	Names = CodeNames
}

func HasWords(name string) bool {
	ld, ok := languages[name]
	return ok && len(ld.WordSets) > 0
}

func HasSnippets(name string) bool {
	ld, ok := languages[name]
	return ok && len(ld.Snippets) > 0
}

func DefaultWordName() string {
	if HasWords("english") {
		return "english"
	}
	if len(WordNames) > 0 {
		return WordNames[0]
	}
	return "english"
}

func DefaultCodeName() string {
	if HasSnippets("go") {
		return "go"
	}
	if len(CodeNames) > 0 {
		return CodeNames[0]
	}
	return DefaultWordName()
}

func WordSetNames(name string) []string {
	ld, ok := languages[name]
	if !ok {
		return nil
	}
	names := make([]string, len(ld.WordSets))
	for i, set := range ld.WordSets {
		names[i] = set.Name
	}
	return names
}

func DefaultWordSet(name string) string {
	sets := WordSetNames(name)
	if len(sets) == 0 {
		return "words"
	}
	return sets[0]
}

func sortWordSets(sets []wordSet) {
	order := map[string]int{"easy": 0, "medium": 1, "hard": 2, "words": 3}
	sort.Slice(sets, func(i, j int) bool {
		oi, iok := order[sets[i].Name]
		oj, jok := order[sets[j].Name]
		if iok && jok {
			return oi < oj
		}
		if iok {
			return true
		}
		if jok {
			return false
		}
		return sets[i].Name < sets[j].Name
	})
}

// RandomWords picks random words for the word-mode typing test.
func RandomWords(name string, setName string, count int) []string {
	words, _ := RandomWordsWithSet(name, setName, count)
	return words
}

// RandomWordsWithSet picks random words and returns the set label that supplied them.
func RandomWordsWithSet(name string, setName string, count int) ([]string, string) {
	ld, ok := languages[name]
	if !ok || len(ld.WordSets) == 0 {
		ld = languages["english"]
	}

	var selected wordSet
	for _, set := range ld.WordSets {
		if set.Name == setName {
			selected = set
			break
		}
	}
	if len(selected.Words) == 0 && len(ld.WordSets) > 0 {
		selected = ld.WordSets[0]
	}

	if len(selected.Words) == 0 {
		return []string{"hello", "world"}, "fallback"
	}

	out := make([]string, count)
	for i := range out {
		idx := rand.Intn(len(selected.Words))
		if i > 0 && len(selected.Words) > 1 && selected.Words[idx] == out[i-1] {
			idx = (idx + 1 + rand.Intn(len(selected.Words)-1)) % len(selected.Words)
		}
		out[i] = selected.Words[idx]
	}
	return out, selected.Name
}

// RandomSnippet picks a random code snippet for code-mode typing.
func RandomSnippet(name string, difficulty string) Snippet {
	ld, ok := languages[name]
	if !ok || len(ld.Snippets) == 0 {
		return Snippet{
			Topic:   "Fallback Words",
			Content: strings.Join(RandomWords(name, difficulty, 50), " "),
		}
	}

	return ld.Snippets[rand.Intn(len(ld.Snippets))]
}

// GetSnippets returns all snippets for a given language
func GetSnippets(name string) []Snippet {
	if ld, ok := languages[name]; ok {
		return ld.Snippets
	}
	return nil
}
