package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"repolens/packer"
	"repolens/tui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func main() {
	// Flags
	nonInteractive := flag.Bool("n", false, "Run in non-interactive mode (packages everything instantly)")
	format := flag.String("f", "markdown", "Output format: 'markdown' or 'xml'")
	compress := flag.Bool("c", true, "Compress code by stripping comments and blank lines")
	dirPath := flag.String("d", ".", "Target directory to package")
	ignoreStr := flag.String("i", "", "Comma-separated list of custom glob patterns to ignore (e.g. '*.json,*.log')")
	outputFile := flag.String("o", "", "Output file path (use '-' for stdout, leave empty for clipboard)")
	maxTokens := flag.Int("s", 0, "Max tokens per output file (default 0 for no limit/splitting)")
	flag.Parse()

	// Check if directory exists
	targetDir, err := filepath.Abs(*dirPath)
	if err != nil {
		fmt.Printf("Error: Invalid directory path: %v\n", err)
		os.Exit(1)
	}

	info, err := os.Stat(targetDir)
	if err != nil || !info.IsDir() {
		fmt.Printf("Error: Target directory does not exist: %s\n", targetDir)
		os.Exit(1)
	}

	// Stylings for summary output
	styleSuccess := lipgloss.NewStyle().Foreground(lipgloss.Color("78")).Bold(true) // Green
	styleWarn := lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Bold(true)   // Yellow/Gold
	styleHighlight := lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Bold(true) // Cyan

	// Parse custom ignores list
	var customIgnores []string
	if *ignoreStr != "" {
		for _, p := range strings.Split(*ignoreStr, ",") {
			trimmed := strings.TrimSpace(p)
			if trimmed != "" {
				customIgnores = append(customIgnores, trimmed)
			}
		}
	}

	if *nonInteractive {
		// Walk directory
		nodes, err := packer.WalkDirectory(targetDir, customIgnores)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error walking directory: %v\n", err)
			os.Exit(1)
		}

		var allWarnings []string
		var parts []string
		var partBuilder strings.Builder
		currentTokens := 0

		// Generate the tree diagram to prepend to every part
		treeDiagram := packer.GenerateTreeDiagram(nodes)
		treeTokens := packer.CalculateTokens(treeDiagram)

		// Start first part with headers
		headerPrefix := "Below is the repository context of the code files packed for analysis.\n\n"
		partBuilder.WriteString(headerPrefix)
		partBuilder.WriteString(treeDiagram)
		currentTokens += packer.CalculateTokens(headerPrefix) + treeTokens

		for _, node := range nodes {
			if node.IsDir || !node.Selected {
				continue
			}

			fullPath := filepath.Join(targetDir, node.Path)
			data, err := os.ReadFile(fullPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", node.Path, err)
				os.Exit(1)
			}
			content := string(data)

			// 1. Secret Scanning
			warnings := packer.DetectSecrets(content, node.Path)
			allWarnings = append(allWarnings, warnings...)

			// 2. Strip Comments if enabled
			if *compress {
				content = packer.StripComments(content, node.Path)
			}

			// Format this single file
			var fileBuilder strings.Builder
			lang := packer.GetMarkdownLang(node.Path)
			sizeKB := float64(len(content)) / 1024.0
			fileTokens := packer.CalculateTokens(content)

			if strings.ToLower(*format) == "xml" {
				safeContent := content
				if strings.Contains(safeContent, "]]>") {
					safeContent = strings.ReplaceAll(safeContent, "]]>", "]]]]><![CDATA[>")
				}
				fileBuilder.WriteString(fmt.Sprintf("<file path=\"%s\" size=\"%.2f KB\" tokens=\"%d\">\n<![CDATA[\n", node.Path, sizeKB, fileTokens))
				fileBuilder.WriteString(safeContent)
				if !strings.HasSuffix(safeContent, "\n") {
					fileBuilder.WriteString("\n")
				}
				fileBuilder.WriteString("]]>\n</file>\n\n")
			} else {
				fileBuilder.WriteString(fmt.Sprintf("## File: %s (Size: %.2f KB | Tokens: %d)\n", node.Path, sizeKB, fileTokens))
				fileBuilder.WriteString("```" + lang + "\n")
				fileBuilder.WriteString(content)
				if !strings.HasSuffix(content, "\n") {
					fileBuilder.WriteString("\n")
				}
				fileBuilder.WriteString("```\n\n")
			}

			fileStr := fileBuilder.String()
			fileTokenCount := packer.CalculateTokens(fileStr)

			// Check if this file would exceed the limit
			if *maxTokens > 0 && currentTokens+fileTokenCount > *maxTokens && partBuilder.Len() > len(headerPrefix)+len(treeDiagram) {
				// Save current part
				parts = append(parts, partBuilder.String())

				// Reset for next part
				partBuilder.Reset()
				partBuilder.WriteString(headerPrefix)
				partBuilder.WriteString(treeDiagram)
				currentTokens = packer.CalculateTokens(headerPrefix) + treeTokens
			}

			partBuilder.WriteString(fileStr)
			currentTokens += fileTokenCount
		}

		// Append the last part if not empty
		if partBuilder.Len() > len(headerPrefix)+len(treeDiagram) {
			parts = append(parts, partBuilder.String())
		}

		// Direct output routing
		if len(parts) == 0 {
			fmt.Fprintln(os.Stderr, "No files were packed.")
			os.Exit(0)
		}

		if *outputFile == "-" {
			// Write directly to stdout (parts separated by marker if multiple)
			for i, part := range parts {
				if len(parts) > 1 {
					fmt.Printf("=== PART %d of %d ===\n", i+1, len(parts))
				}
				fmt.Print(part)
			}
		} else if *outputFile != "" {
			// Write to specified file
			if len(parts) == 1 {
				err = os.WriteFile(*outputFile, []byte(parts[0]), 0644)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
					os.Exit(1)
				}
			} else {
				// Split files
				ext := filepath.Ext(*outputFile)
				base := strings.TrimSuffix(*outputFile, ext)
				for i, part := range parts {
					partPath := fmt.Sprintf("%s_part%d%s", base, i+1, ext)
					err = os.WriteFile(partPath, []byte(part), 0644)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error writing output file %s: %v\n", partPath, err)
						os.Exit(1)
					}
				}
			}
		} else {
			// Copy to clipboard (warn if split, only copy first part)
			err = packer.CopyToClipboard(parts[0])
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error copying to clipboard: %v\n", err)
				os.Exit(1)
			}
		}

		// Print stats to Stderr if printing payload to stdout, otherwise to Stdout
		logWriter := os.Stdout
		if *outputFile == "-" {
			logWriter = os.Stderr
		} else {
			fmt.Fprintln(logWriter, "Packing files in non-interactive mode...")
		}

		// Stats
		totalSize := 0
		totalTokens := 0
		for _, part := range parts {
			totalSize += len(part)
			totalTokens += packer.CalculateTokens(part)
		}

		packedCount := 0
		for _, n := range nodes {
			if !n.IsDir {
				packedCount++
			}
		}

		fmt.Fprintln(logWriter, styleSuccess.Render("✔ Packed successfully!"))
		fmt.Fprintf(logWriter, "• Total packed files: %d\n", packedCount)
		fmt.Fprintf(logWriter, "• Total split parts: %d\n", len(parts))
		fmt.Fprintf(logWriter, "• Estimated total size: %.2f KB\n", float64(totalSize)/1024.0)
		fmt.Fprintf(logWriter, "• Estimated total tokens: %s\n", formatNumber(totalTokens))

		if *outputFile == "" {
			if len(parts) > 1 {
				fmt.Fprintf(logWriter, styleWarn.Render("⚠ Output split into %d parts because it exceeded --max-tokens.\n"), len(parts))
				fmt.Fprintln(logWriter, "👉 Only PART 1 has been copied to your clipboard. Use the -o flag to save all parts to files.")
			} else {
				fmt.Fprintln(logWriter, "📋 Content has been copied to your clipboard.")
			}
		} else if *outputFile != "-" {
			if len(parts) == 1 {
				fmt.Fprintf(logWriter, "💾 Packed content saved to file: %s\n", *outputFile)
			} else {
				ext := filepath.Ext(*outputFile)
				base := strings.TrimSuffix(*outputFile, ext)
				fmt.Fprintf(logWriter, "💾 Packed content saved to files: %s_part1%s to %s_part%d%s\n", base, ext, base, len(parts), ext)
			}
		}

		if len(allWarnings) > 0 {
			fmt.Fprintln(logWriter, styleWarn.Render("\n⚠ Warnings detected during packing:"))
			for _, w := range allWarnings {
				fmt.Fprintf(logWriter, "  - %s\n", w)
			}
		}
		os.Exit(0)
	}

	// Interactive TUI Mode
	m, err := tui.NewModel(targetDir, *format, *compress, customIgnores)
	if err != nil {
		fmt.Printf("Error initializing TUI: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		fmt.Printf("Error running TUI: %v\n", err)
		os.Exit(1)
	}

	// Check results
	mFinal := finalModel.(tui.Model)
	if mFinal.PackingDone {
		totalTokens := packer.CalculateTokens(mFinal.PackedResult)
		packedCount := 0
		for _, n := range mFinal.Nodes {
			if !n.IsDir && n.Selected {
				packedCount++
			}
		}

		fmt.Println()
		fmt.Println(styleSuccess.Render("🎉 RepoLens Packed Successfully!"))
		fmt.Printf("• Files packed: %s\n", styleHighlight.Render(fmt.Sprintf("%d", packedCount)))
		fmt.Printf("• Output format: %s\n", styleHighlight.Render(strings.ToUpper(mFinal.Format)))
		fmt.Printf("• Total tokens copied: %s\n", styleHighlight.Render(formatNumber(totalTokens)))
		fmt.Println("📋 Prompt context is copied to your clipboard. Paste it directly in ChatGPT/Claude.")

		if len(mFinal.Warnings) > 0 {
			fmt.Println(styleWarn.Render("\n⚠ Warnings:"))
			for _, w := range mFinal.Warnings {
				fmt.Printf("  - %s\n", w)
			}
		}
		fmt.Println()
	} else {
		fmt.Println("Packing cancelled.")
	}
}

func formatNumber(n int) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	if n < 1000000 {
		return fmt.Sprintf("%.1fk", float64(n)/1000.0)
	}
	return fmt.Sprintf("%.2fm", float64(n)/1000000.0)
}
