package packer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStripCommentsGo(t *testing.T) {
	input := `package main

// This is a single line comment
import "fmt"

/*
This is a multi-line comment block
*/
func main() {
	fmt.Println("Hello") // Inline comment
}
`
	expected := `package main

import "fmt"

func main() {
	fmt.Println("Hello") 
}
`
	output := StripComments(input, "main.go")
	if strings.TrimSpace(output) != strings.TrimSpace(expected) {
		t.Errorf("Expected stripped Go content to match.\nGot:\n%s\nExpected:\n%s", output, expected)
	}
}

func TestStripCommentsPython(t *testing.T) {
	input := `#!/usr/bin/env python3
# Core logic file
import os

def hello():
    # print hello
    print("hello") # inline comment
`
	expected := `#!/usr/bin/env python3
import os

def hello():
    print("hello") 
`
	output := StripComments(input, "main.py")
	if strings.TrimSpace(output) != strings.TrimSpace(expected) {
		t.Errorf("Expected stripped Python content to match.\nGot:\n%s\nExpected:\n%s", output, expected)
	}
}

func TestDetectSecrets(t *testing.T) {
	tests := []struct {
		filename string
		content  string
		expected bool // true if secret detected
	}{
		{"main.go", "package main\nconst key = \"sk-1234567890abcdef1234567890abcdef1234567890abcdef\"", true},
		{"main.go", "package main\nconst key = \"regular string\"", false},
		{".env", "PORT=8080", true},
		{"config.json", "{\n  \"AWS_KEY\": \"AKIAIOSFODNN7EXAMPLE\"\n}", true},
	}

	for _, tt := range tests {
		warnings := DetectSecrets(tt.content, tt.filename)
		detected := len(warnings) > 0
		if detected != tt.expected {
			t.Errorf("DetectSecrets(%q) on file %q: got detected=%t, want=%t (Warnings: %v)", tt.content, tt.filename, detected, tt.expected, warnings)
		}
	}
}

func TestCalculateTokens(t *testing.T) {
	content := "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"Hello, World!\")\n}"
	tokens := CalculateTokens(content)
	if tokens <= 0 {
		t.Errorf("CalculateTokens returned %d, expected > 0", tokens)
	}
}

func TestWalkDirectory(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "repolens-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Setup directory structure
	// tempDir/
	//   .git/ (should be ignored)
	//     config
	//   src/
	//     main.go
	//     utils.go
	//   .gitignore
	//   README.md

	err = os.MkdirAll(filepath.Join(tempDir, ".git"), 0755)
	if err != nil {
		t.Fatal(err)
	}
	_ = os.WriteFile(filepath.Join(tempDir, ".git", "config"), []byte("[core]"), 0644)

	err = os.MkdirAll(filepath.Join(tempDir, "src"), 0755)
	if err != nil {
		t.Fatal(err)
	}
	_ = os.WriteFile(filepath.Join(tempDir, "src", "main.go"), []byte("package main"), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "src", "utils.go"), []byte("package main"), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, ".gitignore"), []byte("src/utils.go\n"), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "README.md"), []byte("# Test"), 0644)

	nodes, err := WalkDirectory(tempDir, nil)
	if err != nil {
		t.Fatalf("WalkDirectory failed: %v", err)
	}

	// Checked expectations:
	// 1. README.md should be included
	// 2. src/ should be included
	// 3. src/main.go should be included
	// 4. src/utils.go should be ignored (due to gitignore)
	// 5. .git/ should be ignored (due to defaultIgnorePatterns)

	foundReadme := false
	foundSrcDir := false
	foundMainGo := false
	foundUtilsGo := false
	foundGit := false

	for _, n := range nodes {
		switch n.Path {
		case "README.md":
			foundReadme = true
		case "src":
			foundSrcDir = true
		case "src/main.go":
			foundMainGo = true
		case "src/utils.go":
			foundUtilsGo = true
		}
		if n.Path == ".git" || strings.HasPrefix(n.Path, ".git/") {
			foundGit = true
		}
	}

	if !foundReadme {
		t.Error("README.md not found in WalkDirectory output")
	}
	if !foundSrcDir {
		t.Error("src directory not found in WalkDirectory output")
	}
	if !foundMainGo {
		t.Error("src/main.go not found in WalkDirectory output")
	}
	if foundUtilsGo {
		t.Error("src/utils.go was not ignored, but was in .gitignore")
	}
	if foundGit {
		t.Error(".git/ folder was not ignored")
	}
}

func TestFormatPackedOutput(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "repolens-test-format-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	_ = os.WriteFile(filepath.Join(tempDir, "main.go"), []byte("package main\n// comment\nfunc f() {}"), 0644)

	nodes := []*FileNode{
		{
			Path:     "main.go",
			IsDir:    false,
			Selected: true,
		},
	}

	output, _, err := FormatPackedOutput(tempDir, nodes, "markdown", true)
	if err != nil {
		t.Fatalf("FormatPackedOutput failed: %v", err)
	}

	// Should contain language label, file name, and comment stripped content
	if !strings.Contains(output, "## File: main.go") {
		t.Error("Formatted output missing file header")
	}
	if !strings.Contains(output, "```go") {
		t.Error("Formatted output missing markdown block language label")
	}
	if strings.Contains(output, "// comment") {
		t.Error("Formatted output has comment but compression was set to true")
	}

	xmlOutput, _, err := FormatPackedOutput(tempDir, nodes, "xml", true)
	if err != nil {
		t.Fatalf("FormatPackedOutput failed: %v", err)
	}

	if !strings.Contains(xmlOutput, "<file path=\"main.go\"") {
		t.Error("Formatted XML output missing tag")
	}
}

func TestSortOrder(t *testing.T) {
	// Let's verify directory sorting. Directories should come before files in the same directory path hierarchy, then alphabetically.
	// We can walk a constructed temp structure
	tempDir, err := os.MkdirTemp("", "repolens-sort-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	_ = os.WriteFile(filepath.Join(tempDir, "b.txt"), []byte("b"), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "a.txt"), []byte("a"), 0644)
	_ = os.MkdirAll(filepath.Join(tempDir, "c_dir"), 0755)
	_ = os.WriteFile(filepath.Join(tempDir, "c_dir", "sub.txt"), []byte("sub"), 0644)

	nodes, err := WalkDirectory(tempDir, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Expected order:
	// 1. c_dir (directory)
	// 2. c_dir/sub.txt (file inside c_dir)
	// 3. a.txt (file in root, alphabetically first)
	// 4. b.txt (file in root, alphabetically second)
	if len(nodes) != 4 {
		t.Fatalf("Expected 4 nodes, got %d", len(nodes))
	}

	paths := make([]string, len(nodes))
	for i, n := range nodes {
		paths[i] = n.Path
	}

	expectedOrder := []string{"c_dir", "c_dir/sub.txt", "a.txt", "b.txt"}
	for i, p := range expectedOrder {
		if paths[i] != p {
			t.Errorf("Sort order mismatch at index %d: got %s, want %s (Full list: %v)", i, paths[i], p, paths)
		}
	}
}

func TestCustomIgnores(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "repolens-custom-ignore-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	_ = os.WriteFile(filepath.Join(tempDir, "a.json"), []byte("{}"), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "b.txt"), []byte("b"), 0644)
	_ = os.MkdirAll(filepath.Join(tempDir, "logs"), 0755)
	_ = os.WriteFile(filepath.Join(tempDir, "logs", "error.log"), []byte("error"), 0644)

	customIgnores := []string{"*.json", "logs"}
	nodes, err := WalkDirectory(tempDir, customIgnores)
	if err != nil {
		t.Fatal(err)
	}

	// a.json and logs/ should be ignored. Only b.txt should be left.
	for _, n := range nodes {
		if n.Path == "a.json" {
			t.Error("a.json was not ignored by custom ignores (*.json)")
		}
		if strings.HasPrefix(n.Path, "logs") {
			t.Error("logs folder was not ignored by custom ignores")
		}
	}
}

func TestIsBinary(t *testing.T) {
	// Test extension fast path
	if !IsBinary("image.png") {
		t.Error("IsBinary should return true for .png extension")
	}
	if IsBinary("main.go") {
		t.Error("IsBinary should return false for .go extension")
	}

	// Test 512-byte content path
	tempDir, err := os.MkdirTemp("", "repolens-binary-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create a text file
	textFile := filepath.Join(tempDir, "text.dat")
	_ = os.WriteFile(textFile, []byte("Hello world! This is a text file content."), 0644)
	if IsBinary(textFile) {
		t.Error("IsBinary incorrectly flagged text file as binary")
	}

	// Create a binary file (containing null byte)
	binFile := filepath.Join(tempDir, "binary.dat")
	_ = os.WriteFile(binFile, []byte{0x48, 0x65, 0x6C, 0x6C, 0x6F, 0x00, 0x57, 0x6F, 0x72, 0x6C, 0x64}, 0644)
	if !IsBinary(binFile) {
		t.Error("IsBinary failed to detect null-byte binary file")
	}
}

func TestGenerateTreeDiagram(t *testing.T) {
	nodes := []*FileNode{
		{Path: "src", Name: "src", IsDir: true, Depth: 0, Selected: true},
		{Path: "src/main.go", Name: "main.go", IsDir: false, Depth: 1, Selected: true},
		{Path: "README.md", Name: "README.md", IsDir: false, Depth: 0, Selected: true},
		{Path: "secret.key", Name: "secret.key", IsDir: false, Depth: 0, Selected: false}, // Not selected
	}

	diagram := GenerateTreeDiagram(nodes)
	
	if !strings.Contains(diagram, "📁 src/") {
		t.Error("Tree diagram missing directory")
	}
	if !strings.Contains(diagram, "  📄 main.go") {
		t.Error("Tree diagram missing indented file")
	}
	if !strings.Contains(diagram, "📄 README.md") {
		t.Error("Tree diagram missing root file")
	}
	if strings.Contains(diagram, "secret.key") {
		t.Error("Tree diagram should exclude unselected files")
	}
}
