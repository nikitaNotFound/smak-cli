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
	branch internal.Branch
}

func (i branchItem) Title() string {
	return i.branch.Name
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
}

type customDelegate struct {
	list.DefaultDelegate
	selectedIndexes *map[int]bool
}

func (d customDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	
	// Check if this item is marked for deletion
	isMarkedForDeletion := (*d.selectedIndexes)[index]
	isCurrentlyNavigated := index == m.Index()
	
	if isMarkedForDeletion {
		// Override styles for deletion marking
		if isCurrentlyNavigated {
			// Both marked and selected - bold red with underline
			d.DefaultDelegate.Styles.SelectedTitle = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true).Underline(true)
			d.DefaultDelegate.Styles.SelectedDesc = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
		} else {
			// Just marked for deletion - red
			d.DefaultDelegate.Styles.NormalTitle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
			d.DefaultDelegate.Styles.NormalDesc = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
		}
	} else {
		// Reset to default styles
		d.DefaultDelegate.Styles.SelectedTitle = lipgloss.NewStyle().Foreground(lipgloss.Color("170"))
		d.DefaultDelegate.Styles.SelectedDesc = lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
		d.DefaultDelegate.Styles.NormalTitle = lipgloss.NewStyle()
		d.DefaultDelegate.Styles.NormalDesc = lipgloss.NewStyle()
	}
	
	// Use the default delegate's render method
	d.DefaultDelegate.Render(w, m, index, listItem)
}

func newBranchModel(branches []internal.Branch) branchModel {
	items := make([]list.Item, len(branches))
	for i, branch := range branches {
		items[i] = branchItem{branch: branch}
	}
	
	selectedIndexes := make(map[int]bool)
	delegate := customDelegate{
		DefaultDelegate: list.NewDefaultDelegate(),
		selectedIndexes: &selectedIndexes,
	}
	// Set default selection styles
	delegate.DefaultDelegate.Styles.SelectedTitle = lipgloss.NewStyle().Foreground(lipgloss.Color("170"))
	delegate.DefaultDelegate.Styles.SelectedDesc = lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
	
	// Initialize with zero size (will be updated by WindowSizeMsg)
	l := list.New(items, delegate, 0, 0)
	l.Title = "Branches"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	
	return branchModel{
		list:            l,
		branches:        branches,
		selectedIndexes: selectedIndexes,
		confirmDelete:   false,
		helpVisible:     true,
	}
}

func (m branchModel) Init() tea.Cmd {
	return nil
}

func (m branchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := lipgloss.NewStyle().GetFrameSize()
		// Reserve space for help text
		helpHeight := 3
		m.list.SetSize(msg.Width-h, msg.Height-v-helpHeight)
		return m, nil
		
	case tea.KeyMsg:
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
			if len(m.selectedIndexes) > 0 {
				// Clear all selections by clearing the existing map
				for k := range m.selectedIndexes {
					delete(m.selectedIndexes, k)
				}
				return m, nil
			}
			// No selections, quit the program
			return m, tea.Quit
		case "d":
			if len(m.list.Items()) > 0 {
				idx := m.list.Index()
				if m.selectedIndexes[idx] {
					delete(m.selectedIndexes, idx)
				} else {
					m.selectedIndexes[idx] = true
				}
			}
			return m, nil
		case "enter":
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
		if len(m.selectedIndexes) > 0 {
			helpText = "↑↓: navigate • enter: confirm • d: delete • esc: clear • q: quit"
		} else {
			helpText = "↑↓: navigate • enter: checkout • d: delete • esc/q: quit"
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

func init() {
	rootCmd.AddCommand(branchesCmd)
}