package lang

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestParseWordsMonkeytypeObject(t *testing.T) {
	words, label := parseWords("words.json", []byte(`{"name":"test","words":[" alpha ","","beta"]}`))

	want := []string{"alpha", "beta"}
	if !slices.Equal(words, want) {
		t.Fatalf("parseWords() = %#v, want %#v", words, want)
	}
	if label != "test" {
		t.Fatalf("parseWords() label = %q, want %q", label, "test")
	}
}

func TestParseWordsJSONArray(t *testing.T) {
	words, label := parseWords("custom.json", []byte(`["one","two"]`))

	want := []string{"one", "two"}
	if !slices.Equal(words, want) {
		t.Fatalf("parseWords() = %#v, want %#v", words, want)
	}
	if label != "custom" {
		t.Fatalf("parseWords() label = %q, want %q", label, "custom")
	}
}

func TestParseWordsTextUsesFileName(t *testing.T) {
	words, label := parseWords("medium.txt", []byte("one\ntwo one\n"))

	want := []string{"one", "two"}
	if !slices.Equal(words, want) {
		t.Fatalf("parseWords() = %#v, want %#v", words, want)
	}
	if label != "medium" {
		t.Fatalf("parseWords() label = %q, want %q", label, "medium")
	}
}

func TestParseWordsRejectsInvalidUTF8(t *testing.T) {
	words, label := parseWords("bad.txt", []byte{0xff, 0xfe})
	if words != nil || label != "" {
		t.Fatalf("parseWords() = %#v, %q; want nil, empty label", words, label)
	}
}

func TestParseWordsSplitsJSONFields(t *testing.T) {
	words, _ := parseWords("custom.json", []byte(`["one two","two","three"]`))

	want := []string{"one", "two", "three"}
	if !slices.Equal(words, want) {
		t.Fatalf("parseWords() = %#v, want %#v", words, want)
	}
}

func TestLanguageIndexesSplitWordsAndCode(t *testing.T) {
	if !slices.Contains(WordNames, "english") {
		t.Fatal("WordNames does not include english")
	}
	if slices.Contains(CodeNames, "english") {
		t.Fatal("CodeNames includes english word list")
	}
	if !slices.Contains(CodeNames, "go") {
		t.Fatal("CodeNames does not include go")
	}
}

func TestRandomWordsWithSetUsesNamedSet(t *testing.T) {
	_, label := RandomWordsWithSet("english", "easy", 1)
	if label != "easy" {
		t.Fatalf("RandomWordsWithSet() label = %q, want %q", label, "easy")
	}
}

func TestWordSetNames(t *testing.T) {
	sets := WordSetNames("english")
	for _, want := range []string{"easy", "medium", "hard"} {
		if !slices.Contains(sets, want) {
			t.Fatalf("WordSetNames() = %#v, want %q", sets, want)
		}
	}
}

func TestLoadRuntimeLanguages(t *testing.T) {
	oldLanguages := languages
	oldWordNames := WordNames
	oldCodeNames := CodeNames
	oldNames := Names
	t.Cleanup(func() {
		languages = oldLanguages
		WordNames = oldWordNames
		CodeNames = oldCodeNames
		Names = oldNames
	})

	languages = map[string]*langData{}
	WordNames = nil
	CodeNames = nil
	Names = nil

	root := t.TempDir()
	humanDir := filepath.Join(root, "human", "ukrainian")
	if err := os.MkdirAll(humanDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(humanDir, "words.json"), []byte(`{"name":"ukrainian_1k","words":["кіт","пес"]}`), 0644); err != nil {
		t.Fatal(err)
	}

	programmingDir := filepath.Join(root, "programming", "zig")
	if err := os.MkdirAll(programmingDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(programmingDir, "01_basics.zig"), []byte("// Topic: Basics\nconst value = 1;"), 0644); err != nil {
		t.Fatal(err)
	}

	loadRuntime(root)
	finalizeIndexes()

	if !slices.Contains(WordNames, "ukrainian") {
		t.Fatalf("WordNames = %#v, want ukrainian", WordNames)
	}
	if !slices.Equal(WordSetNames("ukrainian"), []string{"ukrainian_1k"}) {
		t.Fatalf("WordSetNames() = %#v, want ukrainian_1k", WordSetNames("ukrainian"))
	}
	if !slices.Contains(CodeNames, "zig") {
		t.Fatalf("CodeNames = %#v, want zig", CodeNames)
	}
	if snippets := GetSnippets("zig"); len(snippets) != 1 || snippets[0].Topic != "Basics" {
		t.Fatalf("GetSnippets() = %#v, want Basics snippet", snippets)
	}
}
