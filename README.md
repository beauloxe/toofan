<div align="center">

# toofan

**A minimal, lightning-fast typing TUI**  
_Practice with local word lists or real code snippets. No browser, no account, everything stays local._

<br>

<img src="assets/main.gif" alt="toofan demo" width="750">

</div>

---

## Differences from upstream

- **UTF-8 Input:** For proper handling of non-ASCII symbol set (e.g. Cyrillic languages).
- **Refactored Language Handling:** Human and programming languages are now split into `internal/lang/data/{human,programming}`, instead of hard-coded `difficulty` there's now `language` and `set` properties.
- **Monkeytype Wordsets:** Added support for Monkeytype [`.json` format](https://github.com/monkeytypegame/monkeytype/tree/master/frontend/static/languages).
- **Runtime Wordset Loading:** Wordsets can now be loaded from `~/.config/toofan/lang/<type>/<language>` dynamically.
- **Nix Flake**: Try it with `nix run github:beauloxe/toofan`!

And some WIP/ideas:
- [ ] Finish dynamic calculations for tables in the profile page
- [ ] Fix duplicate words in typing tests (sometimes two identical words come one after another, e.g. `expensive expensive`)
- [ ] Add `system` theme (use terminal base16 colors)
- [ ] Move to JSON for state data (sorry plain-text enjoyers)

### AI Disclosure

Those changes were implemented using Codex CLI and GPT-5.4/5.5 on Medium thinking. I wanted to add new features for myself because I liked the project, but very little code was written by hand, so no guarantees on code safety, quality, etc.

## Features

- **Two Modes:** Practice human-language word lists or real-world code snippets.
- **Curated Lessons:** Hand-written, topic-based code exercises across multiple languages.
- **Dynamic Themes:** Cycle between multiple aesthetic terminal themes (`ctrl+t`).
- **Live Metrics:** Real-time WPM speed and accuracy tracking.
- **Error Review:** See exactly which words you mistyped after every test.
- **Ranks:** Automated progression system based on your typing speed.
- **Offline & Local:** No browser, no account, zero telemetry.

<p align="center">
  <img src="assets/code-snippets-grid.png" width="48%" title="Real Code Snippets" alt="Real Code Snippets" />
  <img src="assets/lession-grid.png" width="48%" title="Curated Topics & Lessons" alt="Curated Topics & Lessons" />
  <img src="assets/languages-grid.png" width="48%" title="Multiple Languages Supported" alt="Multiple Languages Supported" />
  <img src="assets/theme-grid.png" width="48%" title="Dynamic Built-in Themes" alt="Dynamic Built-in Themes" />
</p>

## Profile Dashboard

A personal overview of your typing speed history, personal bests across durations, and a daily activity map to keep you consistent. Press `ctrl+p` to open.

<div align="center">
<img src="assets/profile-new.png" width="95%">
</div>

## Installation (upstream-only)

⚠️ **Note:** Always take a backup (`ctrl+s`) before updating toofan.

### curl (macOS & Linux)

```bash
curl -fsSL https://raw.githubusercontent.com/vyrx-dev/toofan/master/install.sh | sh
```

### AUR

```bash
paru -S toofan-bin
```

### Go

```bash
go install github.com/vyrx-dev/toofan@latest
```

### Homebrew / Nix / Ubuntu / Fedora

Coming soon.

### Build from Source

If you prefer building manually (requires Go):

```sh
git clone https://github.com/vyrx-dev/toofan.git
cd toofan
go build -o toofan .
mv toofan ~/.local/bin/
```

## FAQ

<details>
<summary>How are stats calculated?</summary>

```text
raw      = total_chars / 5 / elapsed_minutes
wpm      = (total_chars - uncorrected_errors) / 5 / elapsed_minutes
accuracy = (total_chars - all_mistakes) / total_chars × 100
```

- **wpm** - your net speed. Every 5 characters count as one "word". Uncorrected mistakes are subtracted.
- **accuracy** - counts every wrong keystroke, even if you corrected it with backspace.
- **raw** - your gross speed before any penalty.
- **errors** - press `e` on the results page to see exactly which words you mistyped.
</details>

<details>
<summary>How do I change word or code languages?</summary>

Press `ctrl+l` before starting a test. In words mode it shows available word-list languages. In code mode it shows available code-snippet languages.

</details>

<details>
<summary>Where are my files stored?</summary>

Everything lives in `~/.config/toofan/` as plain text files:

- `config.txt` : Your selected duration, mode, language, and theme
- `results.txt` : Every test result (date, wpm, accuracy, duration, mode/language/word set)
- `pb.txt` : Your personal bests per mode and duration
- `lang/` : Optional runtime word lists and code snippets
</details>

<details>
<summary>Can I backup my data?</summary>

Yes. Press `ctrl+s` to save a backup and `ctrl+r` to restore from one. Backups are saved to `~/.config/toofan/backups/` and can be moved between machines.

</details>

<details>
<summary>How do I update toofan?</summary>

The update process depends on how you installed it:

**curl (Quick Install):**
Just run the install command again. It will automatically download and replace the old binary.

```bash
curl -fsSL https://raw.githubusercontent.com/vyrx-dev/toofan/master/install.sh | sh
```

**Go:**

```bash
go install github.com/vyrx-dev/toofan@latest
```

**AUR:**
Use your AUR helper to update the package:

```bash
paru -Syu toofan-bin
```

</details>

<details>
<summary>How do I uninstall Toofan?</summary>

If you installed via the `curl` Quick Install, simply delete the binary and the configuration folder:

```bash
rm ~/.local/bin/toofan
rm -rf ~/.config/toofan
```

_(If you built it from source and moved it globally, run `sudo rm /usr/local/bin/toofan` instead)._

</details>

<details>
<summary>Does it work offline?</summary>

Yes. Everything runs locally and is embedded in the binary. No internet needed.

</details>

<details>
<summary>Want more programming languages?</summary>

We're always looking to add more. If your favorite programming language isn't supported yet, open a PR with a few lesson files and we'll get it in. Check `AGENTS.md` for the file format.

</details>

## Roadmap

- [x] Curl script installation (macOS & Linux)
- [x] Proper documentation for AI and contributors
- [ ] More language support (python, rust, c, typescript, etc.)
- [x] Word set selection for word lists
- [ ] AUR, Homebrew, Nix packages
- [ ] Fix top pane alignment to match bottom panes in profile

## Contributing

- New snippets : Drop a file in `internal/lang/data/programming/<language>/` and rebuild
- New word lists : Add `words.txt` or a Monkeytype-style `words.json` with `{ "name": "...", "words": [...] }`
- New code languages : Add a folder under `internal/lang/data/programming/` with lesson files
- Runtime content : Put files under `~/.config/toofan/lang/human/<language>/` or `~/.config/toofan/lang/programming/<language>/`
- New themes : One Go file with a color palette
- Bug fixes and UX improvements

If you're using an AI coding assistant, read [`AGENTS.md`](AGENTS.md) first.

## Dependencies

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) : TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) : Terminal styling

---

<a href="https://www.star-history.com/#vyrx-dev/toofan&type=date&legend=top-left">
 <picture>
   <source media="(prefers-color-scheme: dark)" srcset="https://api.star-history.com/svg?repos=vyrx-dev/toofan&type=date&theme=dark&legend=top-left" />
   <source media="(prefers-color-scheme: light)" srcset="https://api.star-history.com/svg?repos=vyrx-dev/toofan&type=date&legend=top-left" />
   <img alt="Star History Chart" src="https://api.star-history.com/svg?repos=vyrx-dev/toofan&type=date&legend=top-left" />
 </picture>
</a>
