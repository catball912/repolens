package packer

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Unpack detects the format of the content (XML or Markdown) and writes the files back to the target directory
func Unpack(content string, targetDir string) (int, error) {
	// 1. Try XML parsing
	count, err := UnpackXML(content, targetDir)
	if err != nil {
		return count, err
	}
	if count > 0 {
		return count, nil
	}

	// 2. Try Markdown parsing
	return UnpackMarkdown(content, targetDir)
}

// UnpackXML parses XML-formatted files (wrapped in <file> tags and optional CDATA) and writes them to targetDir
func UnpackXML(content string, targetDir string) (int, error) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	var currentPath string
	var buffer strings.Builder
	inBlock := false
	inCDATA := false
	count := 0

	xmlStartRegex := regexp.MustCompile(`<file\s+path=["']([^"']+)["'][^>]*>`)

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if !inBlock {
			matches := xmlStartRegex.FindStringSubmatch(line)
			if len(matches) > 1 {
				currentPath = matches[1]
				inBlock = true
				inCDATA = false
				buffer.Reset()
			}
		} else {
			// Check if block ends
			if trimmed == "</file>" || (inCDATA && trimmed == "]]>") || (inCDATA && trimmed == "]]>\n</file>") {
				// Strip trailing newline if any
				fileContent := buffer.String()
				if strings.HasSuffix(fileContent, "\n") {
					fileContent = strings.TrimSuffix(fileContent, "\n")
				}
				err := writeUnpackedFile(targetDir, currentPath, fileContent)
				if err != nil {
					return count, err
				}
				count++
				inBlock = false
				inCDATA = false
				continue
			}

			if trimmed == "<![CDATA[" {
				inCDATA = true
				continue
			}

			// Handle inline CDATA closure on the same line if any
			if inCDATA && strings.HasSuffix(trimmed, "]]>") {
				cleaned := strings.TrimSuffix(line, "]]>")
				buffer.WriteString(cleaned)
				fileContent := buffer.String()
				err := writeUnpackedFile(targetDir, currentPath, fileContent)
				if err != nil {
					return count, err
				}
				count++
				inBlock = false
				inCDATA = false
				continue
			}

			// Accumulate content
			buffer.WriteString(line + "\n")
		}
	}
	return count, nil
}

// UnpackMarkdown parses Markdown-formatted code blocks and writes them to targetDir
func UnpackMarkdown(content string, targetDir string) (int, error) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	var currentPath string
	var buffer strings.Builder
	inBlock := false
	inCode := false
	count := 0

	mdHeaderRegex := regexp.MustCompile(`^##\s+File:\s*([^\s\(\)]+)`)

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if !inBlock {
			matches := mdHeaderRegex.FindStringSubmatch(line)
			if len(matches) > 1 {
				currentPath = matches[1]
				inBlock = true
				inCode = false
			}
		} else {
			if !inCode {
				if strings.HasPrefix(trimmed, "```") {
					inCode = true
					buffer.Reset()
				}
			} else {
				if strings.HasPrefix(trimmed, "```") {
					// End of code block
					fileContent := buffer.String()
					if strings.HasSuffix(fileContent, "\n") {
						fileContent = strings.TrimSuffix(fileContent, "\n")
					}
					err := writeUnpackedFile(targetDir, currentPath, fileContent)
					if err != nil {
						return count, err
					}
					count++
					inBlock = false
					inCode = false
					continue
				}
				buffer.WriteString(line + "\n")
			}
		}
	}
	return count, nil
}

func writeUnpackedFile(targetDir string, relPath string, content string) error {
	// Clean path to prevent directory traversal
	cleanedPath := filepath.Clean(relPath)
	if strings.HasPrefix(cleanedPath, "..") || filepath.IsAbs(cleanedPath) {
		return fmt.Errorf("security error: path %q attempts directory traversal outside workspace", relPath)
	}

	fullPath := filepath.Join(targetDir, cleanedPath)

	// Ensure parent directory exists
	dir := filepath.Dir(fullPath)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	// Clean nested CDATA escaping if any (restore "]]>" from "]]]]><![CDATA[>")
	content = strings.ReplaceAll(content, "]]]]><![CDATA[>", "]]>")

	// Write file (overwrite)
	return os.WriteFile(fullPath, []byte(content), 0644)
}