package packer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUnpackXML(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "repolens-unpack-xml-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	xmlInput := `Below is the repository context.

<file path="src/main.go" size="1.2 KB" tokens="350">
<![CDATA[
package main
import "fmt"
func main() {
	fmt.Println("Hello")
}
]]>
</file>

<file path="config.json">
<![CDATA[
{
  "name": "repolens"
}
]]>
</file>
`

	count, err := Unpack(xmlInput, tempDir)
	if err != nil {
		t.Fatalf("Unpack XML failed: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected 2 files unpacked, got %d", count)
	}

	// Verify main.go
	mainBytes, err := os.ReadFile(filepath.Join(tempDir, "src/main.go"))
	if err != nil {
		t.Fatal(err)
	}
	mainStr := string(mainBytes)
	if !strings.Contains(mainStr, "package main") || !strings.Contains(mainStr, "func main()") {
		t.Errorf("Unexpected main.go content:\n%s", mainStr)
	}

	// Verify config.json
	configBytes, err := os.ReadFile(filepath.Join(tempDir, "config.json"))
	if err != nil {
		t.Fatal(err)
	}
	configStr := string(configBytes)
	if !strings.Contains(configStr, `"name": "repolens"`) {
		t.Errorf("Unexpected config.json content:\n%s", configStr)
	}
}

func TestUnpackMarkdown(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "repolens-unpack-md-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	mdInput := `Below is the repository context.

## File: src/main.go (Size: 0.12 KB | Tokens: 35)
` + "```go" + `
package main
import "fmt"
func main() {
	fmt.Println("Hello")
}
` + "```" + `

## File: README.md
` + "```markdown" + `
# RepoLens
This is a markdown README.
` + "```" + `
`

	count, err := Unpack(mdInput, tempDir)
	if err != nil {
		t.Fatalf("Unpack Markdown failed: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected 2 files unpacked, got %d", count)
	}

	// Verify main.go
	mainBytes, err := os.ReadFile(filepath.Join(tempDir, "src/main.go"))
	if err != nil {
		t.Fatal(err)
	}
	mainStr := string(mainBytes)
	if !strings.Contains(mainStr, "package main") {
		t.Errorf("Unexpected main.go content:\n%s", mainStr)
	}

	// Verify README.md
	readmeBytes, err := os.ReadFile(filepath.Join(tempDir, "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	readmeStr := string(readmeBytes)
	if !strings.Contains(readmeStr, "# RepoLens") {
		t.Errorf("Unexpected README.md content:\n%s", readmeStr)
	}
}

func TestUnpackSecurityTraversal(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "repolens-unpack-security-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Traversal attempt via relative path escaping the workspace
	traversalInput := `
<file path="../escaped.txt">
<![CDATA[
this should fail
]]>
</file>
`

	_, err = Unpack(traversalInput, tempDir)
	if err == nil {
		t.Error("Expected error due to directory traversal attempt, but got nil")
	} else if !strings.Contains(err.Error(), "security error") {
		t.Errorf("Expected security error, got: %v", err)
	}
}
