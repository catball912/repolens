# 🔍 RepoLens

> **The Ultimate Local-First LLM Code Packer & Interactive Token Optimizer**

[English](README.md) | [繁體中文](README_ZH.md)

[![Go Report Card](https://goreportcard.com/badge/github.com/catball912/repolens)](https://goreportcard.com/report/github.com/catball912/repolens)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/github/go-mod/go-version/catball912/repolens)](https://github.com/golang/go)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat-package)](http://makeapullrequest.com)

**RepoLens** is a zero-dependency, ultra-fast command-line utility and interactive Terminal User Interface (TUI) that packages your codebase into a single, beautifully structured prompt payload for Large Language Models like **Claude**, **ChatGPT**, and **Gemini**.

Unlike other tools, RepoLens runs entirely locally, estimates tokens in real-time, warns you of hardcoded API keys before you copy, and automatically splits massive repositories into semantic token-limited parts.

![RepoLens TUI Demo](demo.gif)

📖 **Need a detailed guide? Check out our step-by-step tutorials:**
*   👉 **[Detailed Tutorial (English)](TUTORIAL_EN.md)**
*   👉 **[Detailed Tutorial (繁體中文)](TUTORIAL_ZH.md)**

---

## ✨ Features

*   **⚡ Zero Dependencies & Blazing Fast**: Compiled in Go. Run it instantly with a single executable binary. Scans 160k+ tokens in under 8 seconds.
*   **🎮 Premium Interactive TUI**: Elm-architecture terminal interface (built on Charm's Bubble Tea). Navigate your codebase, collapse folders, and selectively check files.
*   **📊 Live Token Counter**: Dynamic progress bar estimates your context window consumption (based on `tiktoken` GPT-4o encoding) on the fly as you toggle files.
*   **🧹 Smart Token Optimization**: Automatically strips single-line and multi-line comments/docstrings for Go, JavaScript, TypeScript, Python, HTML/CSS, and Bash while preserving executable shebangs (`#!/bin/bash`). Saves **15% - 50%** of tokens.
*   **🛡️ Credentials Leak Prevention**: Local scanner flags OpenAI keys, AWS keys, Slack webhooks, and generic passwords in real-time. Highlights files with a `⚠️ [SECRET!]` badge in the TUI explorer.
*   **📁 Prepended Repository Architecture**: Automatically generates and inserts a text-based repository tree diagram at the top of the prompt payload, giving AI models an instant understanding of your codebase directory structure.
*   **📁 Token-based File Splitting**: Specify a token limit (e.g. `-s 50000`) and RepoLens will automatically slice your project into multiple parts, prepending the tree diagram to each part.
*   **🚫 Smart Ignores**: Respects `.gitignore` automatically, skips binary files via a 512-byte null-byte scanner, and supports custom wildcard patterns (e.g., `-i "*_test.go,*.log"`).

---

## 🚀 Quick Start

### 1. Interactive TUI Mode (Best for Humans)
Simply navigate to your repository and launch the interface:
```bash
repolens
```

#### TUI Keyboard Controls
| Key | Action |
| :--- | :--- |
| `↑` / `↓` (or `k` / `j`) | Move explorer cursor up/down |
| `Space` | Select/deselect file (Recursively toggles folders) |
| `←` / `→` (or `h` / `l`) | Collapse/expand folder |
| `c` | Toggle comment stripping / code compression (`ON` / `OFF`) |
| `f` | Toggle format between `MARKDOWN` and `XML` |
| `Enter` | Pack selected files & copy to system clipboard |
| `Esc` / `q` | Exit |

---

### 2. Direct CLI & Bi-directional Mode (Best for Automation & AI Agents)
Package everything instantly, or write mutated code back to your workspace:
```bash
# Package current directory and copy to clipboard
repolens -n

# Package directory, write output to file, and exclude test files
repolens -n -d /path/to/project -i "*_test.go" -o output.md

# Split a massive codebase into files of maximum 50k tokens each
repolens -n -s 50000 -o repo.md

# Apply LLM edits back to code (pipes clipboard response directly back to files!)
pbpaste | repolens -u -

# Unpack LLM changes from a response file into a specific target directory
repolens -d /path/to/project -u response.txt
```

#### CLI Flags
*   `-n`: Run in non-interactive (CLI) mode.
*   `-d <path>`: Target directory to package or unpack into (default: `.`).
*   `-o <path>`: Output file path. Use `-o -` to print directly to stdout, or leave empty to write to the system clipboard.
*   `-s <int>`: Max tokens per split file (default: `0` for no splitting).
*   `-c`: Strip comments and blank lines (default: `true`).
*   `-f <format>`: Output layout style: `markdown` or `xml` (default: `markdown`).
*   `-i <patterns>`: Comma-separated custom ignore glob patterns (e.g. `*.json,*.log`).
*   `-u <path>`: Unpack and write back changes from a packed file or LLM response (use `-` for stdin).

---

## 📊 Real-world Benchmarks
Based on testing against the [spf13/cobra](https://github.com/spf13/cobra) codebase:
*   **Uncompressed prompt**: 621 KB (~168k tokens) in **7.9s**.
*   **Compressed prompt (Comments removed)**: 518 KB (~145k tokens) in **7.9s**.
*   **Average Savings**: **17.2%** of tokens saved per prompt.
*   **Splitting logic**: Splitting the project with `-s 10000` generated 16 files, each complete with a visual directory map and code integrity preserved.

---

## 🛠️ Installation

### Homebrew (macOS / Linux)
*(Coming Soon)*
```bash
brew tap catball912/repolens
brew install repolens
```

### Go Install (Requires Go installed)
```bash
go install github.com/catball912/repolens@latest
```

### Manual Binary Download
Download pre-compiled binaries for macOS, Linux, and Windows from our [Releases Page](https://github.com/catball912/repolens/releases).

---

## 🛡️ Security & Privacy
RepoLens runs **100% locally**. None of your source code, configuration files, or tokens are ever sent to external servers or API hosts. Your secrets stay where they belong: on your machine.

---

## 📄 License
This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
