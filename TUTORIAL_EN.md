# 📖 RepoLens Detailed Tutorial (English Version)

Welcome to **RepoLens**! This step-by-step tutorial will guide you from quick installation and environment path setup to advanced terminal UI (TUI) shortcuts, automation CLI workflows, context token splitting, and privacy security scanning.

---

## 📌 Table of Contents
1. [🛠️ Chapter 1: Installation & PATH Environment Configuration](#1-installation--path-environment-configuration)
2. [🎮 Chapter 2: Interactive TUI Walkthrough (Human Workflow)](#2-interactive-tui-walkthrough-human-workflow)
3. [🤖 Chapter 3: Automation & AI Agent Scripting (Agent Workflow)](#3-automation--ai-agent-scripting-agent-workflow)
4. [🧩 Chapter 4: Large Codebase Splitting Strategy (Token Optimization)](#4-large-codebase-splitting-strategy-token-optimization)
5. [🛡️ Chapter 5: Secrets Leak Prevention & Custom Ignores](#5-secrets-leak-prevention--custom-ignores)

---

## 1. 🛠️ Installation & PATH Environment Configuration

To run `repolens` from **any folder or directory** in your system, you need to configure the binary within your system's `PATH`.

### A. macOS / Linux

#### Option 1: Install via Homebrew (Recommended)
```bash
brew tap catball912/repolens
brew install repolens
```
*Homebrew will automatically handle PATH configurations. Once done, you can run the command immediately.*

#### Option 2: Manual Binary Setup
1. Go to the [Releases Page](https://github.com/catball912/repolens/releases) and download the archive for your architecture (e.g. `repolens-darwin-arm64.tar.gz` for macOS Apple Silicon M1/M2/M3).
2. Unpack the archive to get the `repolens` executable.
3. Open your Terminal and move the binary to a system path (e.g. `/usr/local/bin`):
   ```bash
   sudo mv repolens /usr/local/bin/
   ```
4. Set executable permissions on the binary:
   ```bash
   sudo chmod +x /usr/local/bin/repolens
   ```
5. Try typing `repolens` in any directory to start the tool!

---

### B. Windows
1. Go to the [Releases Page](https://github.com/catball912/repolens/releases) and download `repolens-windows-amd64.zip`.
2. Extract `repolens.exe` and place it in a dedicated folder (e.g., `C:\Program Files\RepoLens\`).
3. Set your system environment variables:
   * Press `Win + R`, type `sysdm.cpl` and hit enter to open System Properties.
   * Go to the **Advanced** tab ➜ click **Environment Variables**.
   * Under "System variables", select **Path** and click **Edit**.
   * Click **New**, and paste the folder path: `C:\Program Files\RepoLens\`.
   * Save and close the dialogs.
4. Open a new PowerShell or CMD window and type `repolens` to run it!

---

## 2. 🎮 Interactive TUI Walkthrough (Human Workflow)

When you run `repolens` in a project, you'll be greeted by an interactive file explorer. This is designed for copy-pasting code context into browser-based Claude or ChatGPT models.

```text
  📁 packer/
    📄 packer.go
  📄 README.md
  
  [c] Strip Comments: ON  |  [f] Format: MARKDOWN  |  Tokens: 1,245 / 200k
```

### ⌨️ Keybindings & Controls
*   **Navigate Cursor**: Use Arrow keys `↑` / `↓` or Vim keys `k` / `j`.
*   **Expand / Collapse Folders**:
    *   On a directory, press `→` (Right Arrow) or `l`: Expand folder.
    *   On a directory, press `←` (Left Arrow) or `h`: Collapse folder.
*   **Select / Deselect**: Press `Space` (Spacebar).
    *   *Tip: Pressing spacebar on a directory recursively checks or unchecks all children inside, making it fast to exclude entire sections (e.g., `tests/`).*
*   **Toggle Compression**: Press `c`.
    *   Default is `ON`. This removes single-line and multi-line comments and blank lines to reduce token size dynamically. You will see the token counter re-estimate instantly.
*   **Toggle Output Format**: Press `f`.
    *   Switch between `MARKDOWN` and `XML` formats.
*   **Pack & Copy**: Press `Enter`.
    *   Instantly packages the selected files and writes them to your system clipboard, then quits to the terminal. Go to Claude and hit `Cmd + V` (paste)!

---

## 3. 🤖 Automation & AI Agent Scripting (Agent Workflow)

For headless operations or scripts where terminal interactions are not possible, RepoLens features a non-interactive CLI mode.

### A. Non-Interactive CLI Command
Use the `-n` flag to pack directories instantly without launching the TUI:
```bash
# Pack current directory and copy to clipboard
repolens -n

# Pack directory, write output to file, and ignore test files
repolens -n -d /home/user/myproject -o packed_project.md
```

### B. Standard Output Redirection (Stdout)
You can direct the packed prompt directly to stdout using `-o -` for integrations. Summary statistics and warnings will automatically redirect to stderr:
```bash
# Package code and feed it directly to Simon Willison's LLM CLI
repolens -n -o - | llm "Explain the architecture of this project"
```

---

## 4. 🧩 Large Codebase Splitting Strategy (Token Optimization)

If a codebase is too large (exceeding 100k or 200k tokens), pasting it all at once can overload the LLM or consume too many message credits. RepoLens supports **Token-based File Splitting**.

### ⚙️ Usage Example
To pack a repository and automatically split the files into parts of maximum 50,000 tokens each:
```bash
repolens -n -s 50000 -o repo_export.md
```

### 💎 Smart Splitting Features:
1.  **File Integrity**: Slices occur cleanly between files. No individual source code file is cut in half.
2.  **Visual Tree Diagrams**: Each generated file (e.g., `repo_export_part1.md`, `repo_export_part2.md`) automatically includes the complete `Repository Tree Structure` diagram at the top. This ensures the LLM always has the overall context of the codebase layout.

---

## 5. 🛡️ Secrets Leak Prevention & Custom Ignores

RepoLens operates **100% locally** to guarantee your private code remains secure. It includes local scanners to prevent accidental leak of API keys or credentials.

### A. Security Scanners
If the pack scanner detects API keys, environment files (`.env`), or passwords:
*   **TUI Mode**: A yellow `⚠️ [SECRET!]` badge will blink next to the file name in the explorer.
*   **CLI Mode**: Stderr logs will output specific warning notifications:
    ```text
    ⚠ Warnings detected during packing:
      - Detected possible OpenAI API Key in config.json
      - Packed a configuration environment (.env) file.
    ```

### B. Custom Ignore Patterns (`-i`)
You can use the `-i` parameter to exclude patterns from being scanned or packed:
```bash
# Ignore test files, nested log files, and configuration files
repolens -n -i "*_test.go,logs/*.log,config.json"
```
By default, RepoLens respects `.gitignore` rules and uses a 512-byte null-byte scanner to skip binary assets (e.g. images, database files, and build outputs) automatically.
