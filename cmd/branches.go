package cmd

import (
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/nikitaNotFound/smak-cli/internal"
)

var branchesCmd = &cobra.Command{
	Use:   "b",
	Short: "Browse and manage branches interactively",
	Long:  `Interactive branch browser with selection, deletion, and navigation features.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := checkGitRepo(); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		branches, err := internal.GetBranches()
		if err != nil {
			fmt.Printf("Error getting branches: %v\n", err)
			return
		}

		model := newBranchModel(branches)
		p := tea.NewProgram(model, tea.WithAltScreen())

		if _, err := p.Run(); err != nil {
			log.Fatalf("Error running program: %v", err)
		}
	},
}

type branchItem struct {
	branch         internal.Branch
	isMergeSource  bool
	isMarkedDelete bool
}

func (i branchItem) Title() string {
	title := i.branch.Name
	if i.isMergeSource {
		title += " (selected to merge from)"
	}
	return title
}

func (i branchItem) Description() string {
	return fmt.Sprintf("%s | %s | ↑%d ↓%d",
		i.branch.LastCommitDate.Format("2006-01-02 15:04"),
		i.branch.LastCommitMessage,
		i.branch.CommitsAhead,
		i.branch.CommitsBehind,
	)
}

func (i branchItem) FilterValue() string {
	return i.branch.Name
}

type branchModel struct {
	list            list.Model
	branches        []internal.Branch
	selectedIndexes map[int]bool
	confirmDelete   bool
	helpVisible     bool
	mergeMode       bool
	mergeSourceIdx  int
	showMergeResult bool
	mergeResult     *internal.MergeResult
	mergeBranches   struct {
		source string
		target string
	}
}

type customDelegate struct {
	list.DefaultDelegate
}

func (d customDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(branchItem)
	if !ok {
		d.DefaultDelegate.Render(w, m, index, listItem)
		return
	}

	isCurrentlyNavigated := index == m.Index()

	// Determine styling based on item state
	var titleStyle, descStyle lipgloss.Style

	if item.isMarkedDelete {
		// Red for deletion
		if isCurrentlyNavigated {
			titleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true).Underline(true)
			descStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
		} else {
			titleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
			descStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
		}
	} else if item.isMergeSource {
		// Orange for merge source
		if isCurrentlyNavigated {
			titleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Bold(true).Underline(true)
			descStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Bold(true)
		} else {
			titleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("208"))
			descStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("208"))
		}
	} else {
		// Default styles
		if isCurrentlyNavigated {
			titleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("170"))
			descStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
		} else {
			titleStyle = lipgloss.NewStyle()
			descStyle = lipgloss.NewStyle()
		}
	}

	fmt.Fprint(w, titleStyle.Render(item.Title()))
	fmt.Fprint(w, "\n")
	fmt.Fprint(w, descStyle.Render(item.Description()))
}

func newBranchModel(branches []internal.Branch) branchModel {
	items := make([]list.Item, len(branches))
	for i, branch := range branches {
		items[i] = branchItem{
			branch:         branch,
			isMergeSource:  false,
			isMarkedDelete: false,
		}
	}

	selectedIndexes := make(map[int]bool)

	// Create model first so we can point to its fields
	m := branchModel{
		branches:        branches,
		selectedIndexes: selectedIndexes,
		confirmDelete:   false,
		helpVisible:     true,
		mergeMode:       false,
		mergeSourceIdx:  -1,
		showMergeResult: false,
		mergeResult:     nil,
	}

	// Create delegate
	delegate := customDelegate{
		DefaultDelegate: list.NewDefaultDelegate(),
	}

	// Initialize with zero size (will be updated by WindowSizeMsg)
	l := list.New(items, delegate, 0, 0)
	l.Title = "Branches"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)

	m.list = l

	return m
}

func (m branchModel) updateListItems() branchModel {
	items := make([]list.Item, len(m.branches))
	for i, branch := range m.branches {
		items[i] = branchItem{
			branch:         branch,
			isMergeSource:  m.mergeMode && i == m.mergeSourceIdx,
			isMarkedDelete: m.selectedIndexes[i],
		}
	}
	m.list.SetItems(items)
	return m
}

func (m branchModel) Init() tea.Cmd {
	return nil
}

func (m branchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if m.showMergeResult {
			// Handle merge result window sizing
			return m, nil
		}
		h, v := lipgloss.NewStyle().GetFrameSize()
		// Reserve space for help text
		helpHeight := 3
		m.list.SetSize(msg.Width-h, msg.Height-v-helpHeight)
		return m, nil

	case tea.KeyMsg:
		if m.showMergeResult {
			switch msg.String() {
			case "enter":
				// Confirm merge or complete successful merge
				if m.mergeResult.Success || !m.mergeResult.HasConflicts {
					// Refresh branches and return to normal mode
					return m.refreshBranchesAndReset()
				}
				// For conflicts, just close the dialog and stay in merge state
				m.showMergeResult = false
				return m, nil
			case "esc":
				// Always abort merge (whether successful or not)
				if err := internal.AbortMerge(); err != nil {
					log.Printf("Error aborting merge: %v", err)
				}
				return m.refreshBranchesAndReset()
			}
			return m, nil
		}

		if m.confirmDelete {
			switch msg.String() {
			case "y", "Y":
				return m.deleteBranches()
			case "n", "N", "esc":
				m.confirmDelete = false
				return m, nil
			}
			return m, nil
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "esc":
			if m.mergeMode {
				// Exit merge mode
				m.mergeMode = false
				m.mergeSourceIdx = -1
				// Update list items to reflect normal state
				m = m.updateListItems()
				return m, nil
			}
			if len(m.selectedIndexes) > 0 {
				// Clear all selections by clearing the existing map
				for k := range m.selectedIndexes {
					delete(m.selectedIndexes, k)
				}
				// Update list items to reflect cleared selections
				m = m.updateListItems()
				return m, nil
			}
			// No selections, quit the program
			return m, tea.Quit
		case "m":
			if !m.mergeMode && len(m.list.Items()) > 0 {
				// Enter merge mode
				m.mergeMode = true
				m.mergeSourceIdx = m.list.Index()
				// Debug: log when entering merge mode
				log.Printf("DEBUG: Entering merge mode, sourceIdx=%d", m.mergeSourceIdx)
				// Update list items to reflect merge state
				m = m.updateListItems()
				return m, nil
			}
			return m, nil
		case "d":
			if !m.mergeMode && len(m.list.Items()) > 0 {
				idx := m.list.Index()
				if m.selectedIndexes[idx] {
					delete(m.selectedIndexes, idx)
				} else {
					m.selectedIndexes[idx] = true
				}
				// Update list items to reflect delete selection state
				m = m.updateListItems()
			}
			return m, nil
		case "enter":
			if m.mergeMode {
				// Perform merge
				if len(m.list.Items()) > 0 {
					targetIdx := m.list.Index()
					if targetIdx != m.mergeSourceIdx && targetIdx < len(m.branches) && m.mergeSourceIdx < len(m.branches) {
						sourceBranch := m.branches[m.mergeSourceIdx].Name
						targetBranch := m.branches[targetIdx].Name

						m.mergeBranches.source = sourceBranch
						m.mergeBranches.target = targetBranch

						result, err := internal.MergeBranch(sourceBranch, targetBranch)
						if err != nil {
							log.Printf("Error performing merge: %v", err)
							return m, nil
						}

						m.mergeResult = result
						m.showMergeResult = true
						return m, nil
					}
				}
				return m, nil
			}

			if len(m.selectedIndexes) > 0 {
				m.confirmDelete = true
				return m, nil
			}
			// No selections, checkout the currently highlighted branch
			if len(m.list.Items()) > 0 {
				idx := m.list.Index()
				if idx < len(m.branches) {
					branchName := m.branches[idx].Name
					err := internal.CheckoutBranch(branchName)
					if err != nil {
						log.Printf("Error checking out branch %s: %v", branchName, err)
					}
				}
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m branchModel) View() string {
	if m.showMergeResult {
		return m.renderMergeResult()
	}

	if m.confirmDelete {
		selectedBranches := []string{}
		for idx := range m.selectedIndexes {
			selectedBranches = append(selectedBranches, m.branches[idx].Name)
		}

		return fmt.Sprintf("\n  Delete branches: %s? (Y/N)\n",
			strings.Join(selectedBranches, ", "))
	}

	view := m.list.View()

	if m.helpVisible {
		helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
		var helpText string
		if m.mergeMode {
			helpText = "↑↓: navigate • enter: merge into selected • esc: exit merge mode • q: quit"
		} else if len(m.selectedIndexes) > 0 {
			helpText = "↑↓: navigate • enter: confirm • d: delete • m: merge • esc: clear • q: quit"
		} else {
			helpText = "↑↓: navigate • enter: checkout • d: delete • m: merge • esc/q: quit"
		}
		help := helpStyle.Render(helpText)
		view += "\n\n" + help
	}

	return view
}

func (m branchModel) deleteBranches() (tea.Model, tea.Cmd) {
	selectedBranches := []string{}
	for idx := range m.selectedIndexes {
		selectedBranches = append(selectedBranches, m.branches[idx].Name)
	}

	if err := internal.DeleteBranches(selectedBranches); err != nil {
		log.Printf("Error deleting branches: %v", err)
		// On error, still refresh but keep going
	}

	// Reload branches from git
	branches, err := internal.GetBranches()
	if err != nil {
		log.Printf("Error reloading branches: %v", err)
		return m, tea.Quit
	}

	// Create new model with updated branches
	newModel := newBranchModel(branches)
	// Preserve window size if we have it
	if m.list.Width() > 0 && m.list.Height() > 0 {
		newModel.list.SetSize(m.list.Width(), m.list.Height())
	}

	return newModel, nil
}

func (m branchModel) refreshBranchesAndReset() (tea.Model, tea.Cmd) {
	// Reload branches from git
	branches, err := internal.GetBranches()
	if err != nil {
		log.Printf("Error reloading branches: %v", err)
		return m, tea.Quit
	}

	// Create new model with updated branches
	newModel := newBranchModel(branches)
	// Preserve window size if we have it
	if m.list.Width() > 0 && m.list.Height() > 0 {
		newModel.list.SetSize(m.list.Width(), m.list.Height())
	}

	return newModel, nil
}

func (m branchModel) renderMergeResult() string {
	if m.mergeResult == nil {
		return "Error: No merge result to display"
	}

	var content []string

	// Title
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("170")).
		Bold(true).
		Padding(1, 2)

	title := fmt.Sprintf("Merge %s → %s", m.mergeBranches.source, m.mergeBranches.target)
	content = append(content, titleStyle.Render(title))

	// Status
	statusStyle := lipgloss.NewStyle().
		Padding(0, 2).
		Margin(1, 0)

	if m.mergeResult.Success {
		successStyle := statusStyle.Copy().Foreground(lipgloss.Color("46"))
		content = append(content, successStyle.Render("✓ Merge completed successfully"))
	} else if m.mergeResult.HasConflicts {
		conflictStyle := statusStyle.Copy().Foreground(lipgloss.Color("196"))
		content = append(content, conflictStyle.Render(fmt.Sprintf("✗ Merge conflicts detected (%d files)", m.mergeResult.ConflictCount)))

		// List conflict files
		if len(m.mergeResult.ConflictFiles) > 0 {
			filesStyle := lipgloss.NewStyle().
				Padding(0, 4).
				Foreground(lipgloss.Color("243"))

			content = append(content, filesStyle.Render("Conflicted files:"))
			for _, file := range m.mergeResult.ConflictFiles {
				content = append(content, filesStyle.Render("• "+file))
			}
		}
	} else {
		errorStyle := statusStyle.Copy().Foreground(lipgloss.Color("196"))
		content = append(content, errorStyle.Render("✗ Merge failed: "+m.mergeResult.ErrorMessage))
	}

	// Help
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Padding(1, 2).
		Margin(1, 0)

	var helpText string
	if m.mergeResult.Success {
		helpText = "enter: continue • esc: abort merge"
	} else if m.mergeResult.HasConflicts {
		helpText = "enter: resolve manually • esc: abort merge"
	} else {
		helpText = "enter: continue • esc: abort merge"
	}

	content = append(content, helpStyle.Render(helpText))

	// Create a bordered container
	containerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("238")).
		Padding(1).
		Margin(2)

	return containerStyle.Render(strings.Join(content, "\n"))
}

func init() {
	rootCmd.AddCommand(branchesCmd)
}
