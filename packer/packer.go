package packer

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/pkoukk/tiktoken-go"
	gitignore "github.com/sabhiram/go-gitignore"
)

// FileNode represents a file or folder in the project tree
type FileNode struct {
	Path       string // Relative path from root, e.g., "src/main.go"
	Name       string // File/Folder name, e.g., "main.go"
	IsDir      bool
	Depth      int
	Collapsed  bool
	Selected   bool
	Size       int64
	TokenCount int  // Cached token count for files
	HasSecret  bool // True if potential secrets are detected
}

// Default ignores if no gitignore is found
var defaultIgnorePatterns = []string{
	".git",
	"node_modules",
	".DS_Store",
	"dist",
	"build",
	"bin",
	"go.sum",
	".gemini",
}

// WalkDirectory reads the directory tree, respects gitignore and custom ignores, and returns sorted FileNodes
func WalkDirectory(root string, customIgnores []string) ([]*FileNode, error) {
	// Clean root path
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	var ignoreMatcher *gitignore.GitIgnore
	gitignorePath := filepath.Join(root, ".gitignore")
	if _, err := os.Stat(gitignorePath); err == nil {
		ignoreMatcher, _ = gitignore.CompileIgnoreFile(gitignorePath)
	}

	var nodes []*FileNode

	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if path == root {
			return nil
		}

		// Calculate relative path
		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		// Normalize slash for cross-platform matching
		matchPath := filepath.ToSlash(relPath)

		// Check custom ignores
		for _, pattern := range customIgnores {
			pattern = strings.TrimSpace(pattern)
			if pattern == "" {
				continue
			}

			// Try matching full relative path
			matched, errMatch := filepath.Match(pattern, matchPath)
			if errMatch == nil && matched {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			// Try matching path segments
			parts := strings.Split(matchPath, "/")
			for _, part := range parts {
				matchedPart, errPart := filepath.Match(pattern, part)
				if errPart == nil && matchedPart {
					if d.IsDir() {
						return filepath.SkipDir
					}
					return nil
				}
			}
		}

		// Check gitignore or default ignores
		if ignoreMatcher != nil {
			if ignoreMatcher.MatchesPath(matchPath) {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		// Apply default ignores as backup/supplement
		for _, pattern := range defaultIgnorePatterns {
			parts := strings.Split(matchPath, "/")
			for _, part := range parts {
				if part == pattern {
					if d.IsDir() {
						return filepath.SkipDir
					}
					return nil
				}
			}
		}

		// Determine depth
		depth := len(strings.Split(matchPath, "/")) - 1

		var size int64
		if !d.IsDir() {
			if IsBinary(path) {
				return nil
			}
			info, err := d.Info()
			if err == nil {
				size = info.Size()
			}
		}

		nodes = append(nodes, &FileNode{
			Path:      matchPath,
			Name:      d.Name(),
			IsDir:     d.IsDir(),
			Depth:     depth,
			Collapsed: false,
			Selected:  true, // Default selected
			Size:      size,
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort nodes using segment-by-segment hierarchical sort (directories first, then files)
	sort.Slice(nodes, func(i, j int) bool {
		partsI := strings.Split(nodes[i].Path, "/")
		partsJ := strings.Split(nodes[j].Path, "/")

		minLen := len(partsI)
		if len(partsJ) < minLen {
			minLen = len(partsJ)
		}

		for k := 0; k < minLen; k++ {
			if partsI[k] != partsJ[k] {
				// Determine if segment k is a directory or file
				isDirI := k < len(partsI)-1 || nodes[i].IsDir
				isDirJ := k < len(partsJ)-1 || nodes[j].IsDir

				if isDirI != isDirJ {
					return isDirI // Directory comes first
				}
				return partsI[k] < partsJ[k]
			}
		}

		return len(partsI) < len(partsJ)
	})

	return nodes, nil
}

// StripComments removes comments and excess empty lines from code
func StripComments(content string, filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	var lines []string
	scanner := bufio.NewScanner(strings.NewReader(content))

	cSingleRegex := regexp.MustCompile(`//.*$`)
	cBlockRegex := regexp.MustCompile(`/\*.*?\*/`)
	pySingleRegex := regexp.MustCompile(`#.*$`)
	inBlockComment := false
	inPyDocstring := false
	pyDocStringChar := ""

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// 1. Handle C-Style Multi-line comments
		if inBlockComment {
			if strings.Contains(line, "*/") {
				parts := strings.SplitN(line, "*/", 2)
				line = parts[1]
				inBlockComment = false
				trimmed = strings.TrimSpace(line)
			} else {
				continue
			}
		}

		// 2. Handle Python Multi-line docstrings
		if inPyDocstring {
			if strings.Contains(line, pyDocStringChar) {
				parts := strings.SplitN(line, pyDocStringChar, 2)
				line = parts[1]
				inPyDocstring = false
				trimmed = strings.TrimSpace(line)
			} else {
				continue
			}
		}

		isCStyle := ext == ".go" || ext == ".js" || ext == ".ts" || ext == ".tsx" || ext == ".jsx" || ext == ".java" || ext == ".c" || ext == ".cpp" || ext == ".h" || ext == ".css"
		
		if isCStyle {
			// Remove inline block comments like /* comment */
			line = cBlockRegex.ReplaceAllString(line, "")
			trimmed = strings.TrimSpace(line)

			// Check for multi-line block comment start
			if strings.Contains(line, "/*") {
				parts := strings.SplitN(line, "/*", 2)
				line = parts[0]
				inBlockComment = true
				trimmed = strings.TrimSpace(line)
			} else if strings.HasPrefix(trimmed, "//") {
				continue
			} else {
				line = cSingleRegex.ReplaceAllString(line, "")
			}
		}

		if ext == ".py" || ext == ".sh" || ext == ".rb" || ext == ".yaml" || ext == ".yml" || ext == ".dockerfile" || ext == "dockerfile" {
			if ext == ".py" {
				// Python docstring detection
				if strings.Contains(line, `"""`) {
					parts := strings.SplitN(line, `"""`, 2)
					if inPyDocstring {
						line = parts[1]
						inPyDocstring = false
					} else {
						// Check if it starts and ends on the same line
						if strings.Contains(parts[1], `"""`) {
							subparts := strings.SplitN(parts[1], `"""`, 2)
							line = parts[0] + subparts[1]
						} else {
							line = parts[0]
							inPyDocstring = true
							pyDocStringChar = `"""`
						}
					}
					trimmed = strings.TrimSpace(line)
				} else if strings.Contains(line, `'''`) {
					parts := strings.SplitN(line, `'''`, 2)
					if inPyDocstring {
						line = parts[1]
						inPyDocstring = false
					} else {
						if strings.Contains(parts[1], `'''`) {
							subparts := strings.SplitN(parts[1], `'''`, 2)
							line = parts[0] + subparts[1]
						} else {
							line = parts[0]
							inPyDocstring = true
							pyDocStringChar = `'''`
						}
					}
					trimmed = strings.TrimSpace(line)
				}
			}

			if strings.HasPrefix(trimmed, "#") {
				if strings.HasPrefix(trimmed, "#!") {
					// Keep shebang intact
				} else {
					continue
				}
			} else {
				line = pySingleRegex.ReplaceAllString(line, "")
			}
		}

		if ext == ".html" || ext == ".xml" || ext == ".svg" {
			if strings.HasPrefix(trimmed, "<!--") && strings.HasSuffix(trimmed, "-->") {
				continue
			}
			re := regexp.MustCompile(`<!--.*?-->`)
			line = re.ReplaceAllString(line, "")
		}

		if strings.TrimSpace(line) != "" {
			lines = append(lines, line)
		} else if strings.TrimSpace(scanner.Text()) == "" {
			if len(lines) > 0 && lines[len(lines)-1] != "" {
				lines = append(lines, "")
			}
		}
	}

	return strings.Join(lines, "\n")
}

// DetectSecrets scans content for potential secrets (e.g., API Keys, AWS keys)
// Returns list of warnings found
func DetectSecrets(content string, filename string) []string {
	var warnings []string

	// Basic check for file extension (e.g., .env)
	if strings.Contains(strings.ToLower(filename), ".env") {
		warnings = append(warnings, "Packed a configuration environment (.env) file.")
	}

	// Regex check for API keys
	rules := map[string]*regexp.Regexp{
		"OpenAI API Key":         regexp.MustCompile(`sk-(proj-)?[a-zA-Z0-9]{32,}`),
		"Generic API Key/Secret": regexp.MustCompile(`(?i)(api_key|apikey|secret_key|secretkey|password|passwd|private_key|db_pass|db_password|credential|token)\s*[:=]\s*["'][a-zA-Z0-9\-_\.\+=]{8,}["']`),
		"Slack Webhook URL":      regexp.MustCompile(`https://hooks\.slack\.com/services/[a-zA-Z0-9_]{8,12}/[a-zA-Z0-9_]{8,12}/[a-zA-Z0-9_]{24}`),
		"AWS Access Key ID":      regexp.MustCompile(`(A3T[A-Z0-9]|AKIA|AGPA|AIDA|AROA|ASCA|ASIA)[A-Z0-9]{16}`),
		"AWS Secret Access Key":  regexp.MustCompile(`(?i)aws_secret[a-zA-Z0-9_]*\s*[:=]\s*["']?[a-zA-Z0-9/\+=]{40}["']?`),
	}

	for name, regex := range rules {
		if regex.MatchString(content) {
			warnings = append(warnings, "Detected possible "+name+" in "+filename)
		}
	}

	return warnings
}

// CalculateTokens estimates the number of tokens in a string
func CalculateTokens(content string) int {
	// Fallback count in case tokenizer fails: ~4 characters per token
	fallbackCount := len(content) / 4

	// We use "o200k_base" encoding as default (GPT-4o)
	tke, err := tiktoken.GetEncoding("o200k_base")
	if err != nil {
		// Fallback to cl100k_base
		tke, err = tiktoken.GetEncoding("cl100k_base")
		if err != nil {
			return fallbackCount
		}
	}

	tokenized := tke.Encode(content, nil, nil)
	return len(tokenized)
}

// GenerateTreeDiagram creates a text visual tree of the selected file structure
func GenerateTreeDiagram(nodes []*FileNode) string {
	var sb strings.Builder
	sb.WriteString("Repository Tree Structure:\n```\n")
	for _, n := range nodes {
		if !n.Selected {
			continue
		}
		indent := strings.Repeat("  ", n.Depth)
		if n.IsDir {
			sb.WriteString(fmt.Sprintf("%s📁 %s/\n", indent, n.Name))
		} else {
			sb.WriteString(fmt.Sprintf("%s📄 %s\n", indent, n.Name))
		}
	}
	sb.WriteString("```\n\n")
	return sb.String()
}

// FormatPackedOutput builds the final prompt payload from selected files
func FormatPackedOutput(root string, nodes []*FileNode, format string, stripComments bool) (string, []string, error) {
	var sb strings.Builder
	var allWarnings []string

	sb.WriteString("Below is the repository context of the code files packed for analysis.\n\n")

	// Prepend directory structure tree diagram
	sb.WriteString(GenerateTreeDiagram(nodes))

	for _, node := range nodes {
		if node.IsDir || !node.Selected {
			continue
		}

		fullPath := filepath.Join(root, node.Path)
		data, err := os.ReadFile(fullPath)
		if err != nil {
			return "", nil, err
		}

		content := string(data)

		// 1. Secret Scanning
		warnings := DetectSecrets(content, node.Path)
		allWarnings = append(allWarnings, warnings...)

		// 2. Strip Comments if enabled
		if stripComments {
			content = StripComments(content, node.Path)
		}

		// 3. Detect file language for markdown blocks
		lang := GetMarkdownLang(node.Path)

		// Calculate size and tokens for metadata header
		sizeKB := float64(len(content)) / 1024.0
		fileTokens := CalculateTokens(content)

		if strings.ToLower(format) == "xml" {
			safeContent := content
			if strings.Contains(safeContent, "]]>") {
				safeContent = strings.ReplaceAll(safeContent, "]]>", "]]]]><![CDATA[>")
			}
			sb.WriteString(fmt.Sprintf("<file path=\"%s\" size=\"%.2f KB\" tokens=\"%d\">\n<![CDATA[\n", node.Path, sizeKB, fileTokens))
			sb.WriteString(safeContent)
			if !strings.HasSuffix(safeContent, "\n") {
				sb.WriteString("\n")
			}
			sb.WriteString("]]>\n</file>\n\n")
		} else {
			// Markdown format
			sb.WriteString(fmt.Sprintf("## File: %s (Size: %.2f KB | Tokens: %d)\n", node.Path, sizeKB, fileTokens))
			sb.WriteString("```" + lang + "\n")
			sb.WriteString(content)
			if !strings.HasSuffix(content, "\n") {
				sb.WriteString("\n")
			}
			sb.WriteString("```\n\n")
		}
	}

	return sb.String(), allWarnings, nil
}

// CopyToClipboard writes the string to system clipboard
func CopyToClipboard(content string) error {
	return clipboard.WriteAll(content)
}

// Helper to map extensions to markdown codeblock labels
func GetMarkdownLang(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".go":
		return "go"
	case ".js", ".jsx":
		return "javascript"
	case ".ts", ".tsx":
		return "typescript"
	case ".py":
		return "python"
	case ".html":
		return "html"
	case ".css":
		return "css"
	case ".json":
		return "json"
	case ".yaml", ".yml":
		return "yaml"
	case ".md":
		return "markdown"
	case ".sh":
		return "bash"
	case ".sql":
		return "sql"
	case ".rs":
		return "rust"
	default:
		return ""
	}
}

// IsBinary detects if a file is binary by looking at its extension and scanning the first 512 bytes for null bytes
func IsBinary(filePath string) bool {
	// 1. Check extension first (fast path)
	ext := strings.ToLower(filepath.Ext(filePath))
	binaryExts := map[string]bool{
		".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".ico": true,
		".zip": true, ".tar": true, ".gz": true, ".7z": true, ".rar": true,
		".pdf": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true, ".ppt": true, ".pptx": true,
		".exe": true, ".dll": true, ".so": true, ".dylib": true, ".bin": true, ".out": true,
		".woff": true, ".woff2": true, ".ttf": true, ".eot": true,
		".mp3": true, ".mp4": true, ".avi": true, ".mkv": true, ".mov": true,
		".db": true, ".sqlite": true,
	}
	if binaryExts[ext] {
		return true
	}

	// 2. Read first 512 bytes and check for null byte
	file, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer file.Close()

	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err.Error() != "EOF" {
		return false
	}

	for i := 0; i < n; i++ {
		if buf[i] == 0x00 {
			return true
		}
	}

	return false
}
