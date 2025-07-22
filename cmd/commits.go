package cmd

import (
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	
	"github.com/nikitaNotFound/smak-cli/internal"
)

var commitsCmd = &cobra.Command{
	Use:   "c",
	Short: "Browse commits in current branch",
	Long:  `Interactive commit browser with diff viewer.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := checkGitRepo(); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		
		commits, err := internal.GetCommits()
		if err != nil {
			fmt.Printf("Error getting commits: %v\n", err)
			return
		}
		
		model := newCommitModel(commits)
		p := tea.NewProgram(model, tea.WithAltScreen())
		
		if _, err := p.Run(); err != nil {
			log.Fatalf("Error running program: %v", err)
		}
	},
}

type commitItem struct {
	commit internal.Commit
}

func (i commitItem) Title() string {
	shortHash := i.commit.Hash
	if len(shortHash) > 8 {
		shortHash = shortHash[:8]
	}
	return fmt.Sprintf("%s - %s", shortHash, i.commit.Message)
}

func (i commitItem) Description() string {
	return fmt.Sprintf("%s by %s", i.commit.Date.Format("2006-01-02 15:04:05"), i.commit.Author)
}

func (i commitItem) FilterValue() string {
	return i.commit.Message
}

type commitModel struct {
	list        list.Model
	commits     []internal.Commit
	viewport    viewport.Model
	showDiff    bool
	currentDiff string
	helpVisible bool
}

func newCommitModel(commits []internal.Commit) commitModel {
	items := make([]list.Item, len(commits))
	for i, commit := range commits {
		items[i] = commitItem{commit: commit}
	}
	
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().Foreground(lipgloss.Color("170"))
	delegate.Styles.SelectedDesc = lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
	
	l := list.New(items, delegate, 0, 0)
	l.Title = "Commits"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	
	vp := viewport.New(0, 0)
	vp.Style = lipgloss.NewStyle().Border(lipgloss.RoundedBorder())
	
	return commitModel{
		list:        l,
		commits:     commits,
		viewport:    vp,
		showDiff:    false,
		helpVisible: true,
	}
}

func (m commitModel) Init() tea.Cmd {
	return nil
}

func (m commitModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if m.showDiff {
			m.viewport.Width = msg.Width - 2
			m.viewport.Height = msg.Height - 6
		} else {
			h, v := lipgloss.NewStyle().GetFrameSize()
			helpHeight := 3
			m.list.SetSize(msg.Width-h, msg.Height-v-helpHeight)
		}
		return m, nil
		
	case tea.KeyMsg:
		if m.showDiff {
			switch msg.String() {
			case "q", "ctrl+c", "esc":
				m.showDiff = false
				return m, nil
			case "up", "k":
				m.viewport.LineUp(1)
				return m, nil
			case "down", "j":
				m.viewport.LineDown(1)
				return m, nil
			case "pgup", "b":
				m.viewport.ViewUp()
				return m, nil
			case "pgdown", "f":
				m.viewport.ViewDown()
				return m, nil
			}
			return m, nil
		}
		
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			if len(m.list.Items()) > 0 {
				idx := m.list.Index()
				commit := m.commits[idx]
				diff, err := internal.GetCommitDiff(commit.Hash)
				if err != nil {
					return m, nil
				}
				m.currentDiff = diff
				m.viewport.SetContent(diff)
				m.showDiff = true
				return m, nil
			}
		}
	}
	
	if !m.showDiff {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}
	
	return m, nil
}

func (m commitModel) View() string {
	if m.showDiff {
		commit := m.commits[m.list.Index()]
		header := lipgloss.NewStyle().
			Foreground(lipgloss.Color("170")).
			Bold(true).
			Render(fmt.Sprintf("Commit: %s", commit.Hash))
		
		author := lipgloss.NewStyle().
			Foreground(lipgloss.Color("243")).
			Render(fmt.Sprintf("Author: %s", commit.Author))
		
		date := lipgloss.NewStyle().
			Foreground(lipgloss.Color("243")).
			Render(fmt.Sprintf("Date: %s", commit.Date.Format("2006-01-02 15:04:05")))
		
		message := lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			Render(fmt.Sprintf("Message: %s", commit.Message))
		
		help := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Render("↑↓/j k: scroll • pgup/pgdown: page • esc: back • q: quit")
		
		headerContent := lipgloss.JoinVertical(lipgloss.Left, header, author, date, message, "")
		
		return lipgloss.JoinVertical(lipgloss.Left,
			headerContent,
			m.viewport.View(),
			"",
			help,
		)
	}
	
	view := m.list.View()
	
	if m.helpVisible {
		helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
		help := helpStyle.Render("↑↓: navigate • enter: view commit diff • q: quit")
		view += "\n\n" + help
	}
	
	return view
}

func init() {
	rootCmd.AddCommand(commitsCmd)
}