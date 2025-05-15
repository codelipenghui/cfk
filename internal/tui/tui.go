// Package tui provides the terminal user interface for the cfk application
package tui

import (
	"fmt"

	"github.com/cfk-dev/cfk/internal/config"
	"github.com/cfk-dev/cfk/internal/kafka"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model represents the TUI application model
type Model struct {
	config      *config.AppConfig
	kafkaClient *kafka.Client
	state       string
	topicList   list.Model
	topicTable  table.Model
	viewport    viewport.Model
	err         error
}

// Initialize the TUI model
func NewModel(cfg *config.AppConfig) Model {
	// Create topic list
	topicList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	topicList.Title = "Topics"

	// Create topic table
	columns := []table.Column{
		{Title: "Topic", Width: 30},
		{Title: "Partitions", Width: 10},
		{Title: "Replicas", Width: 10},
	}
	rows := []table.Row{}
	topicTable := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	// Create viewport for message viewing
	viewport := viewport.New(80, 20)
	viewport.Style = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder())

	return Model{
		config:     cfg,
		state:      "clusters",
		topicList:  topicList,
		topicTable: topicTable,
		viewport:   viewport,
	}
}

// Init initializes the TUI model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles TUI events and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	// Update components based on current state
	switch m.state {
	case "topics":
		m.topicList, cmd = m.topicList.Update(msg)
		return m, cmd
	case "topic_details":
		m.topicTable, cmd = m.topicTable.Update(msg)
		return m, cmd
	case "messages":
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	default: // clusters
		// For now, just handle basic navigation
		return m, cmd
	}
}

// View renders the TUI
func (m Model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n\nPress any key to exit.", m.err)
	}

	switch m.state {
	case "topics":
		return m.topicList.View()
	case "topic_details":
		return m.topicTable.View()
	case "messages":
		return m.viewport.View()
	default: // clusters
		return "Welcome to cfk - Console for Kafka!\n\n" +
			"Press 'q' to quit.\n"
	}
}

// Start starts the TUI application
func Start(cfg *config.AppConfig) error {
	p := tea.NewProgram(NewModel(cfg), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
