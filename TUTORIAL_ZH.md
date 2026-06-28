# 📖 RepoLens 繁體中文使用教程 (Detailed Tutorial)

歡迎使用 **RepoLens**！本教程將手把手帶您從零開始安裝、設定環境變數，並深入探索互動式 TUI、自動化 CLI 腳本整合、大型專案分卷打包，以及隱私安全防護等高階功能。

---

## 📌 目錄
1. [🛠️ 第一章：安裝與環境變數 PATH 配置](#1-安裝與環境變數-path-配置)
2. [🎮 第二章：互動式 TUI 介面詳解 (人類協作流)](#2-互動式-tui-介面詳解-人類協作流)
3. [🤖 第三章：自動化與 AI Agent 腳本整合 (Agent 協作流)](#3-自動化與-ai-agent-腳本整合-agent-協作流)
4. [🧩 第四章：超大型專案分卷打包策略 (Token Optimization)](#4-超大型專案分卷打包策略-token-optimization)
5. [🛡️ 第五章：隱私安全掃描與自定義無視規則](#5-隱私安全掃描與自定義無視規則)

---

## 1. 🛠️ 安裝與環境變數 PATH 配置

為了讓您能在**任何資料夾/專案目錄**下直接輸入 `repolens` 啟動工具，您需要將二進位檔案放到系統的 `PATH` 中。

### A. macOS / Linux 系統

#### 方法一：使用 Homebrew 安裝 (推薦)
```bash
brew tap catball912/repolens
brew install repolens
```
*Homebrew 會自動幫您配置好環境變數，安裝完成後即可直接使用。*

#### 方法二：手動下載二進位檔案
1. 前往 [Releases 頁面](https://github.com/catball912/repolens/releases) 下載適用於您系統的壓縮包（例如 macOS M系列晶片下載 `repolens-darwin-arm64.tar.gz`）。
2. 解壓檔案取得 `repolens` 執行檔。
3. 打開終端機，將其移動到系統的執行路徑（例如 `/usr/local/bin`）：
   ```bash
   sudo mv repolens /usr/local/bin/
   ```
4. 賦予檔案執行權限：
   ```bash
   sudo chmod +x /usr/local/bin/repolens
   ```
5. 現在，在任意目錄輸入 `repolens` 即可運行！

---

### B. Windows 系統
1. 前往 [Releases 頁面](https://github.com/catball912/repolens/releases) 下載 `repolens-windows-amd64.zip`。
2. 解壓檔案取得 `repolens.exe`，將其放入一個專用資料夾（例如 `C:\Program Files\RepoLens\`）。
3. 配置環境變數：
   * 按 `Win + R` 輸入 `sysdm.cpl` 打開系統內容。
   * 點選 **進階** 頁籤 ➜ **環境變數**。
   * 在「系統變數」列表中找到 **Path**，按 **編輯**。
   * 點選 **新增**，輸入剛才的資料夾路徑 `C:\Program Files\RepoLens\`。
   * 一路點選確定保存。
4. 打開新的 PowerShell 或 CMD，輸入 `repolens` 即可啟動！

---

## 2. 🎮 互動式 TUI 介面詳解 (人類協作流)

當您在終端機輸入 `repolens` 啟動後，會進入一個互動式選單。這非常適合您在編寫代碼時，手動挑選檔案貼給網頁版 Claude 或 ChatGPT。

```text
  📁 packer/
    📄 packer.go
  📄 README.md
  
  [c] Strip Comments: ON  |  [f] Format: MARKDOWN  |  Tokens: 1,245 / 200k
```

### ⌨️ TUI 實戰操作指南
*   **游標移動**：使用鍵盤的 `↑` / `↓` 箭頭或 Vim 鍵位 `k` / `j`。
*   **展開與摺疊目錄**：
    *   在目錄名稱上按 `→` (右箭頭) 或 `l`：展開該目錄。
    *   在目錄名稱上按 `←` (左箭頭) 或 `h`：摺疊該目錄。
*   **遞迴選擇/反選**：按 `Space` (空白鍵)。
    *   *提示：在目錄上按空白鍵，會自動勾選或取消該目錄下的所有子檔案，方便快速剔除整個資料夾（例如 `tests/`）。*
*   **切換代碼壓縮模式**：按 `c` 鍵。
    *   預設為 `ON`，會移除所有單行/多行註釋和空行。您會看到右下角的 Token 估計值即時重新計算並減少。
*   **切換輸出格式**：按 `f` 鍵。
    *   可在 `MARKDOWN` 和 `XML` 之間切換。
*   **一鍵複製並退出**：按 `Enter` 鍵。
    *   選好需要的檔案後，按 `Enter`。RepoLens 會瞬間將代碼排版好並複製到您的剪貼簿中。您可以直接去 Claude 按 `Cmd + V` (貼上)。

---

## 3. 🤖 自動化與 AI Agent 腳本整合 (Agent 協作流)

如果您正在使用一些命令行 AI 工具，或者在編寫自動化腳本，RepoLens 提供了強大的非交互式 CLI 模式。

### A. 基礎 CLI 命令
使用 `-n` 旗標開啟非交互模式（立即執行打包，不顯示 TUI）：
```bash
# 打包當前專案並直接複製到剪貼簿
repolens -n

# 打包指定專案目錄並輸出為本地檔案
repolens -n -d /home/user/myproject -o packed_project.md
```

### B. 導向標準輸出 (Stdout) 供 AI 管道使用
如果您在編寫自動化腳本或讓 AI Agent 調用，您可以使用 `-o -` 將打包好的代碼直接輸出到 stdout，這時日誌與警告會被自動分流到 stderr：
```bash
# 打包並直接將結果傳遞給 Simon Willison 的 LLM CLI 進行提問
repolens -n -o - | llm "請幫我分析這個專案的架構"
```

---

## 4. 🧩 超大型專案分卷打包策略 (Token Optimization)

當您的專案非常大（例如超過 10 萬甚至 20 萬 Tokens），一次性貼給 AI 會導致超出上下文上限或費用過於昂貴。RepoLens 提供了同類工具中最強大的**自動分卷功能**。

### ⚙️ 實戰案例
假設您想將專案打包，但限制每個包最大只能有 50,000 Tokens：
```bash
repolens -n -s 50000 -o repo_export.md
```

### 💎 分卷特色與優勢：
1.  **檔案邊界完整**：RepoLens 不會在檔案的文字中間暴力切斷，而是以檔案為單位進行安全分割，確保代碼結構完整。
2.  **自動 prepended 架構樹**：生成的多個檔案（如 `repo_export_part1.md`、`repo_export_part2.md`）中，**每一個檔案的最頂部都會自動附帶完整的 Repository Tree Structure 目錄圖**，讓 AI 無論讀到哪一個 Part，都能掌握全局專案架構。

---

## 5. 🛡️ 隱私安全掃描與自定義無視規則

RepoLens 承諾 **100% 本地安全執行**。我們內建了敏感資訊檢測器，防止您不小心將 API Keys 或密碼發送給第三方 AI 平台。

### A. 敏感資訊檢測警告
當打包過程中掃描到疑似 API Key、環境變量檔（`.env`）、AWS 密鑰或硬編碼密碼時：
*   **TUI 互動模式**：會在該檔案旁邊亮起黃色的 `⚠️ [SECRET!]` 警示標章，提醒您取消勾選。
*   **CLI 非交互模式**：會將詳細的警告警告日誌輸出到終端：
    ```text
    ⚠ Warnings detected during packing:
      - Detected possible OpenAI API Key in config.json
      - Packed a configuration environment (.env) file.
    ```

### B. 自定義無視規則 (`-i` / `--ignore`)
您可以利用 `-i` 參數來手動排除包含敏感資訊或無用的檔案（支持萬用字元 `*`）：
```bash
# 排除所有測試檔、日誌檔及配置文件
repolens -n -i "*_test.go,logs/*.log,config.json"
```

此外，RepoLens 會**自動讀取並遵守您專案中的 `.gitignore` 檔案**，並利用 512 字節掃描法自動跳過所有圖片、資料庫和編譯好的二進位檔案。
