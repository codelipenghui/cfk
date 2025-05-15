package tui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// CustomDelegate is a custom delegate for the list
type CustomDelegate struct {
	styles *list.DefaultItemStyles
}

// NewCustomDelegate creates a new custom delegate
func NewCustomDelegate() list.ItemDelegate {
	styles := list.NewDefaultItemStyles()
	return &CustomDelegate{
		styles: &styles,
	}
}

// Height returns the height of the delegate
func (d CustomDelegate) Height() int {
	return 1
}

// Spacing returns the spacing of the delegate
func (d CustomDelegate) Spacing() int {
	return 0
}

// Update updates the delegate
func (d CustomDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

// Render renders the delegate
func (d CustomDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(Item)
	if !ok {
		fmt.Fprintf(w, "Error: item is not of type Item")
		return
	}

	title := item.Title()
	desc := item.Description()

	// Get the correct style for the item
	var style lipgloss.Style
	if index == m.Index() {
		style = d.styles.SelectedTitle
	} else {
		style = d.styles.NormalTitle
	}

	// Render the item
	fmt.Fprintf(w, "%s %s", style.Render(title), d.styles.NormalDesc.Render(desc))
}