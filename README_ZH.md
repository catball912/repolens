# 🔍 RepoLens

> **極致的本地優先 LLM 代碼打包與互動式 Token 優化器**

[English](README.md) | [繁體中文](README_ZH.md)

[![Go Report Card](https://goreportcard.com/badge/github.com/catball912/repolens)](https://goreportcard.com/report/github.com/catball912/repolens)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/github/go-mod/go-version/catball912/repolens)](https://github.com/golang/go)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat-package)](http://makeapullrequest.com)

**RepoLens** 是一款無依賴、極速的命令列工具和互動式終端介面（TUI），可將您的代碼庫打包成一個結構清晰的 Prompt Payload，直接餵給 **Claude**、**ChatGPT** 和 **Gemini** 等大型語言模型。

與其他同類工具不同，RepoLens 完全在本地運行、即時估算 Token、在複製前主動掃描並警告硬編碼的 API 金鑰，並能自動將大型專案進行 Token 分卷打包。

📖 **需要詳細指南？請參閱我們的：**
*   👉 **[繁體中文詳細教程 (TUTORIAL_ZH.md)](TUTORIAL_ZH.md)**
*   👉 **[英文詳細教程 (TUTORIAL_EN.md)](TUTORIAL_EN.md)**

---

## ✨ 特色功能

*   **⚡ 零依賴與極致效能**：基於 Go 語言編譯，單一執行檔開箱即用。僅需不到 8 秒即可掃描並處理超過 16 萬個 Token。
*   **🎮 高級互動式 TUI**：基於 Charm 的 Bubble Tea 框架構建。支持在終端機中上下導覽、展開/摺疊目錄、手動挑選打包檔案。
*   **📊 即時 Token 計數器**：動態進度條在您勾選檔案時，會即時估算當前的 Context 佔用量（基於 GPT-4o 的 `tiktoken` 編碼）。
*   **🧹 智能 Token 壓縮**：自動為 Go、JS、TS、Python、HTML/CSS 和 Bash 去除單行/多行註釋和空行，同時完整保留 Shell 腳本的 Shebang 標頭（`#!/bin/bash`），節省 **15% - 50%** 的 Token。
*   **🛡️ 敏感資訊洩漏防護**：本地掃描器在打包前自動識別 OpenAI Key、AWS 密鑰、Slack Webhooks 和通用密碼。在 TUI 中會直接在問題檔案旁亮起 `⚠️ [SECRET!]` 黃色警告。
*   **📁 自動目錄樹導覽**：打包結果的最頂部會自動插入以 📁/📄 表示的專案目錄樹，讓 AI 模型能瞬間建立對程式碼架構的全局認知。
*   **🧩 Token 分卷打包**：為大專案提供分卷功能（例如使用 `-s 50000`）。自動切分檔案並為每一卷的頭部加上專案目錄樹，確保 LLM 連續閱讀。
*   **🚫 智能忽略規則**：自動讀取並遵守 `.gitignore`，利用 512 字節掃描法自動過濾二進位檔案（如圖片、資料庫），並支援自定義 Glob 排除規律（例如 `-i "*_test.go,*.log"`）。

---

## 🚀 快速開始

### 1. 互動式 TUI 模式 (適合開發者手動挑選)
在您的專案根目錄下直接執行：
```bash
repolens
```

#### TUI 鍵盤操作說明
| 按鍵 | 操作 |
| :--- | :--- |
| `↑` / `↓` (或 `k` / `j`) | 移動檔案樹游標 |
| `Space` (空白鍵) | 選擇/取消選取檔案（在目錄上按會遞迴勾選所有子檔案） |
| `←` / `→` (or `h` / `l`) | 摺疊 / 展開目錄 |
| `c` | 切換代碼壓縮（註釋過濾） `ON` / `OFF` |
| `f` | 切換輸出排版格式： `MARKDOWN` / `XML` |
| `Enter` | 開始打包所選檔案，自動寫入系統剪貼簿並退出 |
| `Esc` / `q` | 退出 |

---

### 2. CLI 模式 (適合自動化腳本與 AI Agent 呼叫)
無需啟動選單，直接在後台完成打包：
```bash
# 直接打包當前專案並寫入剪貼簿
repolens -n

# 打包指定專案、排除測試檔並寫入 Markdown 檔案
repolens -n -d /path/to/project -i "*_test.go" -o output.md

# 自動分卷打包：限制每卷最大 50k tokens，輸出多個檔案
repolens -n -s 50000 -o repo.md
```

#### CLI 常用 Flag 參數
*   `-n`：啟用非交互式（CLI）模式。
*   `-d <path>`：目標打包路徑（預設：`.`）。
*   `-o <path>`：輸出檔案路徑。使用 `-o -` 直接輸出到 Stdout，留空則複製到系統剪貼簿。
*   `-s <int>`：每卷最大 Token 上限（預設 `0` 代表不限制/不分卷）。
*   `-c`：是否過濾註釋和空行（預設 `true`）。
*   `-f <format>`：輸出格式類型：`markdown` 或 `xml`（預設 `markdown`）。
*   `-i <patterns>`：以逗號分隔的自定義排除 Glob 規律（如 `*.json,*.log`）。

---

## 📊 真實專案基準測試
針對 [spf13/cobra](https://github.com/spf13/cobra) 進行測試：
*   **未壓縮打包**：**621 KB** (~168k tokens)，用時 **7.96秒**。
*   **壓縮打包 (移除註釋)**：**518 KB** (~145k tokens)，用時 **7.91秒**。
*   **Token 節省率**：**17.2%**。
*   **分卷測試**：使用 `-s 10000` 限制，自動分拆成 16 個獨立檔案，各分卷均保留目錄樹並完美保持了程式碼完整性。

---

## 🛠️ 安裝方式

### Homebrew (macOS / Linux)
*(即將推出)*
```bash
brew tap catball912/repolens
brew install repolens
```

### Go Install 安裝 (需要 Go 環境)
```bash
go install github.com/catball912/repolens@latest
```

### 手動下載執行檔
直接前往 [Releases 頁面](https://github.com/catball912/repolens/releases) 下載適用於 macOS、Linux、Windows 的預編譯二進位執行檔。

---

## 🛡️ 安全與隱私
RepoLens **100% 在本地運行**。您的代碼、設定檔或任何 API Key 絕不會被傳送到外部伺服器。隱私只留在您的主機上。

---

## 📄 開源授權
本專案採用 MIT 授權協議 - 詳情請參閱 [LICENSE](LICENSE) 檔案。
