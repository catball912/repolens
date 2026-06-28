package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"repolens/packer"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Msg types for Bubble Tea
type TokenCalcMsg struct {
	Path   string
	Tokens int
}

type TokenCalcDoneMsg struct{}

type StatusMsg struct {
	Text string
}

// Model represents the TUI state
type Model struct {
	Root           string
	Nodes          []*packer.FileNode
	Cursor         int // Cursor position in the VISIBLE nodes list
	Format         string
	StripComments  bool
	PackingDone    bool
	Warnings       []string
	PackedResult   string
	TotalFiles     int
	FilesCalculated int
	Calclating      bool

	// Lipgloss styles
	styleCursor       lipgloss.Style
	styleDir          lipgloss.Style
	styleFile         lipgloss.Style
	styleCheck        lipgloss.Style
	styleUncheck      lipgloss.Style
	styleTitle        lipgloss.Style
	styleStatusBar    lipgloss.Style
	styleProgressBg   lipgloss.Style
	styleProgressFg   lipgloss.Style
	styleSummaryTitle lipgloss.Style
}

// NewModel initializes the Bubble Tea model
func NewModel(root string, format string, stripComments bool, customIgnores []string) (Model, error) {
	nodes, err := packer.WalkDirectory(root, customIgnores)
	if err != nil {
		return Model{}, err
	}

	totalFiles := 0
	for _, n := range nodes {
		if !n.IsDir {
			totalFiles++
		}
	}

	// Setup visual styles using Lipgloss
	m := Model{
		Root:           root,
		Nodes:          nodes,
		Format:         format,
		StripComments:  stripComments,
		TotalFiles:     totalFiles,
		Calclating:     true,
		Cursor:         0,

		styleCursor:       lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Bold(true), // Cyan
		styleDir:          lipgloss.NewStyle().Foreground(lipgloss.Color("99")).Bold(true), // Light Violet/Purple
		styleFile:         lipgloss.NewStyle().Foreground(lipgloss.Color("252")),          // Off-white
		styleCheck:        lipgloss.NewStyle().Foreground(lipgloss.Color("78")).Bold(true), // Green [x]
		styleUncheck:      lipgloss.NewStyle().Foreground(lipgloss.Color("243")),          // Gray [ ]
		styleTitle:        lipgloss.NewStyle().Background(lipgloss.Color("99")).Foreground(lipgloss.Color("255")).Padding(0, 1).Bold(true),
		styleStatusBar:    lipgloss.NewStyle().Background(lipgloss.Color("236")).Foreground(lipgloss.Color("246")).Padding(0, 1),
		styleProgressBg:   lipgloss.NewStyle().Background(lipgloss.Color("238")),
		styleProgressFg:   lipgloss.NewStyle().Background(lipgloss.Color("86")),
		styleSummaryTitle: lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Bold(true), // Gold/Yellow
	}

	return m, nil
}

// Init starts token calculations asynchronously
func (m Model) Init() tea.Cmd {
	return RecalculateTokensCmd(m.Root, m.Nodes, m.StripComments)
}

// TokensCalculatedMsg contains updated counts and secret warnings
type TokensCalculatedMsg struct {
	TokenCounts map[string]int
	SecretPaths map[string]bool
}

// RecalculateTokensCmd handles parsing tokens for all files in background
func RecalculateTokensCmd(root string, nodes []*packer.FileNode, stripComments bool) tea.Cmd {
	return func() tea.Msg {
		counts := make(map[string]int)
		secrets := make(map[string]bool)
		for _, n := range nodes {
			if n.IsDir {
				continue
			}
			fullPath := filepath.Join(root, n.Path)
			data, err := os.ReadFile(fullPath)
			if err != nil {
				continue
			}
			content := string(data)
			
			// Check for secrets
			warnings := packer.DetectSecrets(content, n.Path)
			if len(warnings) > 0 {
				secrets[n.Path] = true
			}

			if stripComments {
				content = packer.StripComments(content, n.Path)
			}
			counts[n.Path] = packer.CalculateTokens(content)
		}
		return TokensCalculatedMsg{TokenCounts: counts, SecretPaths: secrets}
	}
}

// getVisibleNodes filters out nodes that reside in collapsed parent directories
func (m Model) getVisibleNodes() []*packer.FileNode {
	var visible []*packer.FileNode
	collapsedPaths := make(map[string]bool)

	for _, n := range m.Nodes {
		if n.IsDir && n.Collapsed {
			collapsedPaths[n.Path] = true
		}
	}

	for _, n := range m.Nodes {
		hidden := false
		parts := strings.Split(n.Path, "/")
		for i := 1; i < len(parts); i++ {
			parentPath := strings.Join(parts[:i], "/")
			if collapsedPaths[parentPath] {
				hidden = true
				break
			}
		}

		if !hidden {
			visible = append(visible, n)
		}
	}

	return visible
}

// Update handles state changes from messages and keypresses
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case TokensCalculatedMsg:
		m.Calclating = false
		for _, n := range m.Nodes {
			if count, ok := msg.TokenCounts[n.Path]; ok {
				n.TokenCount = count
			}
			if hasSec, ok := msg.SecretPaths[n.Path]; ok {
				n.HasSecret = hasSec
			} else {
				n.HasSecret = false
			}
		}
		m.FilesCalculated = m.TotalFiles
		return m, nil

	case tea.KeyMsg:
		visible := m.getVisibleNodes()
		if len(visible) == 0 {
			return m, nil
		}

		// Keep cursor in bounds
		if m.Cursor >= len(visible) {
			m.Cursor = len(visible) - 1
		}
		if m.Cursor < 0 {
			m.Cursor = 0
		}

		selectedNode := visible[m.Cursor]

		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit

		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}

		case "down", "j":
			if m.Cursor < len(visible)-1 {
				m.Cursor++
			}

		case "right", "l":
			// Expand directory
			if selectedNode.IsDir && selectedNode.Collapsed {
				selectedNode.Collapsed = false
			}

		case "left", "h":
			// Collapse directory
			if selectedNode.IsDir && !selectedNode.Collapsed {
				selectedNode.Collapsed = true
			}

		case " ":
			// Toggle selected state of node and recursively apply to children
			targetState := !selectedNode.Selected
			selectedNode.Selected = targetState

			if selectedNode.IsDir {
				prefix := selectedNode.Path + "/"
				for _, n := range m.Nodes {
					if strings.HasPrefix(n.Path, prefix) {
						n.Selected = targetState
					}
				}
			}

		case "c", "C":
			// Toggle comment stripping
			m.StripComments = !m.StripComments
			// Recalculate tokens asynchronously to reflect change
			m.Calclating = true
			m.FilesCalculated = 0
			return m, RecalculateTokensCmd(m.Root, m.Nodes, m.StripComments)

		case "f", "F":
			// Toggle format between markdown and xml
			if m.Format == "markdown" {
				m.Format = "xml"
			} else {
				m.Format = "markdown"
			}

		case "enter":
			// Execute packing
			m.PackingDone = true
			result, warnings, err := packer.FormatPackedOutput(m.Root, m.Nodes, m.Format, m.StripComments)
			if err != nil {
				m.Warnings = append(m.Warnings, "Error packing files: "+err.Error())
				return m, tea.Quit
			}

			m.PackedResult = result
			m.Warnings = append(m.Warnings, warnings...)

			// Copy to clipboard
			err = packer.CopyToClipboard(result)
			if err != nil {
				m.Warnings = append(m.Warnings, "Failed to copy to clipboard: "+err.Error())
			}

			return m, tea.Quit
		}
	}

	return m, nil
}

// View renders the terminal screen
func (m Model) View() string {
	var sb strings.Builder

	// Header / Title
	sb.WriteString("\n")
	sb.WriteString(m.styleTitle.Render(" RepoLens v1.0 • Interactive Code Packer "))
	sb.WriteString("\n\n")

	// Calculations status
	if m.Calclating {
		sb.WriteString(fmt.Sprintf(" ⏳ Calculating tokens: %d/%d files...\n", m.FilesCalculated, m.TotalFiles))
	} else {
		sb.WriteString(" ✅ Token calculations complete.\n")
	}

	// Calculate totals
	totalSize := int64(0)
	totalTokens := 0
	selectedCount := 0
	totalVisibleFiles := 0

	for _, n := range m.Nodes {
		if n.IsDir {
			continue
		}
		totalVisibleFiles++
		if n.Selected {
			selectedCount++
			totalSize += n.Size
			totalTokens += n.TokenCount
		}
	}

	// Display summary stats
	sb.WriteString(fmt.Sprintf(" Selected: %d/%d files | Size: %.2f KB | Estimated Tokens: %s\n\n",
		selectedCount, totalVisibleFiles, float64(totalSize)/1024.0, formatNumber(totalTokens)))

	// Scroll viewport limits
	visible := m.getVisibleNodes()
	height := 16 // Number of visible tree lines we want to render
	start := 0
	if m.Cursor >= height {
		start = m.Cursor - height + 1
	}
	end := start + height
	if end > len(visible) {
		end = len(visible)
	}

	// Render file tree
	for i := start; i < end; i++ {
		node := visible[i]

		// Cursor marker
		cursorMarker := "  "
		isCursor := i == m.Cursor
		if isCursor {
			cursorMarker = m.styleCursor.Render("▸ ")
		}

		// Checkbox
		checkbox := ""
		if node.Selected {
			checkbox = m.styleCheck.Render("[x]")
		} else {
			checkbox = m.styleUncheck.Render("[ ]")
		}

		// Indentation & Expand/Collapse indicators
		indent := strings.Repeat("  ", node.Depth)
		expander := ""
		if node.IsDir {
			if node.Collapsed {
				expander = "▶ "
			} else {
				expander = "▼ "
			}
		}

		// Icon & Name style
		nameStr := node.Name
		if node.IsDir {
			nameStr = m.styleDir.Render(nameStr)
		} else {
			nameStr = m.styleFile.Render(nameStr)
		}

		// Token count badge for files
		badge := ""
		if !node.IsDir && node.TokenCount > 0 {
			badge = fmt.Sprintf(" (%s tkn)", formatNumber(node.TokenCount))
			badge = m.styleUncheck.Render(badge) // styled in gray
		}

		// Secret warning badge
		secretBadge := ""
		if node.HasSecret {
			secretBadge = m.styleSummaryTitle.Render(" ⚠️ [SECRET!]")
		}

		// Compile line
		line := fmt.Sprintf("%s%s %s %s%s%s%s\n", cursorMarker, checkbox, indent, expander, nameStr, badge, secretBadge)
		if isCursor {
			// Highlight background for cursor line
			line = lipgloss.NewStyle().Background(lipgloss.Color("237")).Render(line[:len(line)-1]) + "\n"
		}
		sb.WriteString(line)
	}

	// Empty lines to preserve terminal layout height
	for i := len(visible); i < height; i++ {
		sb.WriteString("\n")
	}

	// Progress bar representation
	progressBarWidth := 50
	limit := 200000 // Standard Claude / GPT-4o limit context size
	pct := float64(totalTokens) / float64(limit)
	if pct > 1.0 {
		pct = 1.0
	}
	filled := int(pct * float64(progressBarWidth))
	empty := progressBarWidth - filled

	barStr := m.styleProgressFg.Render(strings.Repeat("█", filled)) + m.styleProgressBg.Render(strings.Repeat("░", empty))
	sb.WriteString(fmt.Sprintf(" Context Fit: [ %s ] %.1f%% of 200k Limit\n\n", barStr, pct*100))

	// Configuration bar
	compLabel := "Off"
	if m.StripComments {
		compLabel = "On"
	}
	sb.WriteString(fmt.Sprintf(" Format: %s | Strip Comments: %s\n", strings.ToUpper(m.Format), compLabel))

	// Footer / Status Bar
	statusBarContent := " [Space] Toggle • [←/→] Collapse/Expand • [C] Strip Comments • [F] Format • [Enter] Copy & Close "
	sb.WriteString(m.styleStatusBar.Render(statusBarContent) + "\n")

	return sb.String()
}

// Helpers
func formatNumber(n int) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	if n < 1000000 {
		return fmt.Sprintf("%.1fk", float64(n)/1000.0)
	}
	return fmt.Sprintf("%.2fm", float64(n)/1000000.0)
}
