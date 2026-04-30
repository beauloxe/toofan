# AGENTS.md

Project context for AI coding assistants working on this repository.

## Overview

Toofan is a terminal typing test written in Go. It uses Bubble Tea for the TUI and Lipgloss for styling. Everything runs offline — words, code snippets, themes, and user data are all local.

## Build & Run

```sh
go build -o toofan .   # compile binary
go run .               # run without building
go vet ./...           # lint check before committing
```

There are no external services, databases, or network calls. User data lives in `~/.config/toofan/` as plain text files.

## Project Structure

```
main.go                         Entry point. Initializes Bubble Tea program.
internal/
  game/
    game.go                     Core game logic: typing, backspace, stats, timer.
    storage.go                  All file I/O: config, results, PBs, backups.
  tui/
    model.go                    Bubble Tea model, Update loop, View dispatch.
    typing.go                   Typing screen: input handling, live WPM, progress bar.
    results.go                  Post-test results screen.
    profile.go                  Profile dashboard: overview, PBs, ranks, history, heatmap.
    text.go                     Text rendering: syntax coloring, word wrapping, cursor.
    picker.go                   Picker overlays: language, lesson, theme, duration.
  lang/
    lang.go                     Loads embedded word lists and code snippets from data/.
    data/human/<language>/      One directory per human word-list language.
      words.txt                 Word list (one word per line).
      words.json                Monkeytype-style word list: { "name": "...", "words": [...] }.
    data/programming/<language>/ One directory per programming language.
      <topic>.<ext>             Code lessons (hand-written, not generated).
  theme/
    theme.go                    Palette struct, All slice, Next/ByName helpers.
    <name>.go                   One file per theme. Each exports a single Palette var.
```

## Architecture Rules

- `game/` has zero UI dependencies. It must never import `tui/`, `theme/`, or `lipgloss`.
- `tui/` reads game state through accessors (`g.Text()`, `g.Input()`, `g.Errors()`). It never touches game internals directly.
- All file I/O is centralized in `storage.go`. No file reads/writes anywhere else.
- `lang/` uses `embed.FS` to bake data files into the binary at compile time. No runtime file system access for content.

## Adding Content

### New Theme

Create `internal/theme/<name>.go`:

```go
package theme

import "github.com/charmbracelet/lipgloss"

var MyTheme = Palette{
    Name:       "mytheme",
    Background: lipgloss.Color("#1a1a1a"),
    Foreground: lipgloss.Color("#555555"),
    Typed:      lipgloss.Color("#ffffff"),
    Error:      lipgloss.Color("#ff0000"),
    Cursor:     lipgloss.Color("#ffffff"),
    Accent:     lipgloss.Color("#00ff00"),
    Success:    lipgloss.Color("#00ff00"),
}
```

Then add `MyTheme` to the `All` slice in `theme.go`.

### New Language

1. Create `internal/lang/data/programming/<language>/` directory.
2. Add lesson files inside it. Each file is a self-contained code snippet.
3. Every lesson file must start with a `// Topic: <title>` comment line. The comment prefix depends on the language (`//`, `#`, or `--`).
4. The code below the comments is what the user types.
5. Rebuild. The `embed.FS` picks up new files automatically.

### New Word List

Place a word file inside `internal/lang/data/human/<language>/`. Supported formats:

- `words.txt`, one word per line or whitespace-separated.
- `easy.txt`, `medium.txt`, `hard.txt` for named word sets.
- `words.json` or any `.json` Monkeytype word list with `{ "name": "...", "words": [...] }`.
- Top-level JSON string arrays are also accepted.

The profile stats use the JSON `name` as the word-set label when present. Plain text files use the filename without extension.

Runtime user content is also loaded from `~/.config/toofan/lang/` with the same split:

- `~/.config/toofan/lang/human/<language>/words.json`
- `~/.config/toofan/lang/programming/<language>/<topic>.<ext>`

## Coding Conventions

- No unnecessary comments. Code should be self-explanatory.
- No emoji in source code or CLI output.
- Error handling: return early, don't wrap in else blocks.
- Naming: short, lowercase Go conventions. No stuttering (`game.GameState` → `game.State`).
- Keep functions short. If a view function exceeds ~60 lines, extract helpers.
- Lipgloss styles are created inline where used. No global style variables.

## Data Format

Results are stored in `~/.config/toofan/results.txt`, one test per line:

```
2026-04-01 22:18 |  85 wpm | 97.5% | 30s | words:english:english_1k | 83 raw | 2 err
2026-04-01 22:20 |  54 wpm | 91.0% | 15s | code:go | 60 raw | 5 err
```

Config is stored in `~/.config/toofan/config.txt` as key=value pairs:

```
duration=30
mode=words
lang=english
wordset=easy
theme=tokyonight
```

PBs are stored in `~/.config/toofan/pb.txt`:

```
words-30=105
code-15=54
```

## Do Not

- Do not add external API calls or network requests.
- Do not add dependencies beyond Bubble Tea and Lipgloss.
- Do not use AI-generated code snippets for lessons. All lessons are hand-written.
- Do not modify the data format in `results.txt` or `pb.txt` without updating both `storage.go` and `profile.go` parsers.
- Do not add global mutable state outside of `theme.Current`.
