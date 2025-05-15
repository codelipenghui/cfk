package tui

import (
	"fmt"
	"strings"

	"github.com/cfk-dev/cfk/internal/config"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ClusterForm represents a form for adding/editing a cluster
type ClusterForm struct {
	inputs       []textinput.Model
	focusIndex   int
	submitButton string
	cancelButton string
	buttonFocus  int // 0 for submit, 1 for cancel
	width        int
	height       int
	cluster      *config.KafkaClusterConfig
	isEdit       bool
}

// NewClusterForm creates a new cluster form
func NewClusterForm(width, height int, cluster *config.KafkaClusterConfig) ClusterForm {
	isEdit := cluster != nil
	
	// Initialize with default values if not editing
	if !isEdit {
		cluster = &config.KafkaClusterConfig{
			Name:      "",
			Bootstrap: []string{""},
			SSL:       false,
			SASL:      false,
		}
	}
	
	// Create form inputs
	inputs := make([]textinput.Model, 4)
	
	// Name input
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Cluster Name"
	inputs[0].Focus()
	inputs[0].Width = 30
	inputs[0].Prompt = "› "
	inputs[0].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	inputs[0].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	inputs[0].SetValue(cluster.Name)
	
	// Bootstrap servers input
	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Bootstrap Servers (comma-separated)"
	inputs[1].Width = 40
	inputs[1].Prompt = "› "
	inputs[1].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	inputs[1].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	inputs[1].SetValue(strings.Join(cluster.Bootstrap, ","))
	
	// Username input
	inputs[2] = textinput.New()
	inputs[2].Placeholder = "Username (optional)"
	inputs[2].Width = 30
	inputs[2].Prompt = "› "
	inputs[2].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	inputs[2].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	inputs[2].SetValue(cluster.Username)
	
	// Password input
	inputs[3] = textinput.New()
	inputs[3].Placeholder = "Password (optional)"
	inputs[3].Width = 30
	inputs[3].Prompt = "› "
	inputs[3].EchoMode = textinput.EchoPassword
	inputs[3].EchoCharacter = '•'
	inputs[3].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	inputs[3].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	inputs[3].SetValue(cluster.Password)
	
	submitText := "Add"
	if isEdit {
		submitText = "Update"
	}
	
	return ClusterForm{
		inputs:       inputs,
		focusIndex:   0,
		submitButton: submitText,
		cancelButton: "Cancel",
		width:        width,
		height:       height,
		cluster:      cluster,
		isEdit:       isEdit,
	}
}

// Init initializes the form
func (f ClusterForm) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles form events
func (f ClusterForm) Update(msg tea.Msg) (ClusterForm, tea.Cmd) {
	var cmds []tea.Cmd
	
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "shift+tab", "up", "down":
			// Cycle focus between inputs and buttons
			if msg.String() == "up" || msg.String() == "shift+tab" {
				// Move focus up
				if f.focusIndex == -1 {
					// If we're on buttons, move to last input
					f.focusIndex = len(f.inputs) - 1
					f.buttonFocus = -1
				} else if f.focusIndex > 0 {
					// Move to previous input
					f.focusIndex--
				} else {
					// Wrap to buttons
					f.focusIndex = -1
					f.buttonFocus = 1
				}
			} else {
				// Move focus down
				if f.focusIndex == -1 {
					// If we're on buttons, move to first input or cycle between buttons
					if f.buttonFocus < 1 {
						f.buttonFocus++
					} else {
						// Wrap to first input
						f.buttonFocus = -1
						f.focusIndex = 0
					}
				} else if f.focusIndex < len(f.inputs)-1 {
					// Move to next input
					f.focusIndex++
				} else {
					// Move to buttons
					f.focusIndex = -1
					f.buttonFocus = 0
				}
			}
			
			// Update focus states
			for i := 0; i < len(f.inputs); i++ {
				if i == f.focusIndex {
					cmds = append(cmds, f.inputs[i].Focus())
				} else {
					f.inputs[i].Blur()
				}
			}
			
			return f, tea.Batch(cmds...)
			
		case "enter":
			if f.focusIndex == -1 {
				// Button is focused
				if f.buttonFocus == 0 {
					// Submit button
					return f, func() tea.Msg {
						// Create cluster config from form data
						bootstrapServers := strings.Split(f.inputs[1].Value(), ",")
						for i, server := range bootstrapServers {
							bootstrapServers[i] = strings.TrimSpace(server)
						}
						
						cluster := config.KafkaClusterConfig{
							Name:      f.inputs[0].Value(),
							Bootstrap: bootstrapServers,
							Username:  f.inputs[2].Value(),
							Password:  f.inputs[3].Value(),
							SSL:       f.cluster.SSL,
							SASL:      f.cluster.SASL,
							SASLType:  f.cluster.SASLType,
						}
						
						if f.isEdit {
							return ClusterUpdatedMsg{Cluster: cluster}
						}
						return ClusterAddedMsg{Cluster: cluster}
					}
				} else {
					// Cancel button
					return f, func() tea.Msg {
						return ClusterFormCancelledMsg{}
					}
				}
			}
		}
	}
	
	// Handle character input for the focused input
	if f.focusIndex >= 0 {
		var cmd tea.Cmd
		f.inputs[f.focusIndex], cmd = f.inputs[f.focusIndex].Update(msg)
		return f, cmd
	}
	
	return f, nil
}

// View renders the form
func (f ClusterForm) View() string {
	var formTitle string
	if f.isEdit {
		formTitle = "Edit Kafka Cluster"
	} else {
		formTitle = "Add New Kafka Cluster"
	}
	
	formStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1, 2).
		Width(f.width - 4)
	
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		MarginBottom(1)
	
	inputsView := ""
	for i, input := range f.inputs {
		var label string
		switch i {
		case 0:
			label = "Cluster Name:"
		case 1:
			label = "Bootstrap Servers:"
		case 2:
			label = "Username (optional):"
		case 3:
			label = "Password (optional):"
		}
		
		labelStyle := lipgloss.NewStyle().Width(20)
		inputsView += labelStyle.Render(label) + " " + input.View() + "\n\n"
	}
	
	// Render buttons
	submitBgColor := "240"
	if f.buttonFocus == 0 {
		submitBgColor = "205"
	}
	
	cancelBgColor := "240"
	if f.buttonFocus == 1 {
		cancelBgColor = "205"
	}
	
	submitButtonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Background(lipgloss.Color(submitBgColor)).
		Padding(0, 3).
		MarginRight(1)
	
	cancelButtonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Background(lipgloss.Color(cancelBgColor)).
		Padding(0, 3)
	
	buttonsView := submitButtonStyle.Render(f.submitButton) + " " + cancelButtonStyle.Render(f.cancelButton)
	
	helpText := "\nUse tab/shift+tab to navigate, enter to submit"
	
	return formStyle.Render(
		titleStyle.Render(formTitle) + "\n" +
			inputsView +
			buttonsView +
			helpText,
	)
}

// ClusterAddedMsg is sent when a new cluster is added
type ClusterAddedMsg struct {
	Cluster config.KafkaClusterConfig
}

// ClusterUpdatedMsg is sent when a cluster is updated
type ClusterUpdatedMsg struct {
	Cluster config.KafkaClusterConfig
}

// ClusterFormCancelledMsg is sent when the form is cancelled
type ClusterFormCancelledMsg struct{}

// TopicForm represents a form for adding/editing a topic
type TopicForm struct {
	inputs       []textinput.Model
	focusIndex   int
	submitButton string
	cancelButton string
	buttonFocus  int // 0 for submit, 1 for cancel
	width        int
	height       int
	topicName    string
	isEdit       bool
}

// NewTopicForm creates a new topic form
func NewTopicForm(width, height int, topicName string, partitions int) TopicForm {
	isEdit := topicName != ""
	
	// Create form inputs
	inputs := make([]textinput.Model, 2)
	
	// Topic name input
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Topic Name"
	inputs[0].Focus()
	inputs[0].Width = 30
	inputs[0].Prompt = "› "
	inputs[0].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	inputs[0].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	inputs[0].SetValue(topicName)
	
	// Partitions input
	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Number of Partitions"
	inputs[1].Width = 10
	inputs[1].Prompt = "› "
	inputs[1].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	inputs[1].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	
	if isEdit {
		inputs[1].SetValue(fmt.Sprintf("%d", partitions))
	} else {
		inputs[1].SetValue("1") // Default to 1 partition
	}
	
	submitText := "Add"
	if isEdit {
		submitText = "Update"
	}
	
	return TopicForm{
		inputs:       inputs,
		focusIndex:   0,
		submitButton: submitText,
		cancelButton: "Cancel",
		width:        width,
		height:       height,
		topicName:    topicName,
		isEdit:       isEdit,
	}
}

// Init initializes the form
func (f TopicForm) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles form events
func (f TopicForm) Update(msg tea.Msg) (TopicForm, tea.Cmd) {
	var cmds []tea.Cmd
	
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "shift+tab", "up", "down":
			// Cycle focus between inputs and buttons
			if msg.String() == "up" || msg.String() == "shift+tab" {
				// Move focus up
				if f.focusIndex == -1 {
					// If we're on buttons, move to last input
					f.focusIndex = len(f.inputs) - 1
					f.buttonFocus = -1
				} else if f.focusIndex > 0 {
					// Move to previous input
					f.focusIndex--
				} else {
					// Wrap to buttons
					f.focusIndex = -1
					f.buttonFocus = 1
				}
			} else {
				// Move focus down
				if f.focusIndex == -1 {
					// If we're on buttons, move to first input or cycle between buttons
					if f.buttonFocus < 1 {
						f.buttonFocus++
					} else {
						// Wrap to first input
						f.buttonFocus = -1
						f.focusIndex = 0
					}
				} else if f.focusIndex < len(f.inputs)-1 {
					// Move to next input
					f.focusIndex++
				} else {
					// Move to buttons
					f.focusIndex = -1
					f.buttonFocus = 0
				}
			}
			
			// Update focus states
			for i := 0; i < len(f.inputs); i++ {
				if i == f.focusIndex {
					cmds = append(cmds, f.inputs[i].Focus())
				} else {
					f.inputs[i].Blur()
				}
			}
			
			return f, tea.Batch(cmds...)
			
		case "enter":
			if f.focusIndex == -1 {
				// Button is focused
				if f.buttonFocus == 0 {
					// Submit button
					return f, func() tea.Msg {
						// Parse partitions
						partitions := 1
						if f.inputs[1].Value() != "" {
							_, err := fmt.Sscanf(f.inputs[1].Value(), "%d", &partitions)
							if err != nil || partitions < 1 {
								return ErrorMsg{err: fmt.Errorf("invalid number of partitions")}
							}
						}
						
						if f.isEdit {
							return TopicUpdatedMsg{
								OldName:    f.topicName,
								Name:       f.inputs[0].Value(),
								Partitions: partitions,
							}
						}
						return TopicAddedMsg{
							Name:              f.inputs[0].Value(),
							Partitions:        partitions,
							ReplicationFactor: 1, // Default to 1 replica
						}
					}
				} else {
					// Cancel button
					return f, func() tea.Msg {
						return TopicFormCancelledMsg{}
					}
				}
			}
		}
	}
	
	// Handle character input for the focused input
	if f.focusIndex >= 0 {
		var cmd tea.Cmd
		f.inputs[f.focusIndex], cmd = f.inputs[f.focusIndex].Update(msg)
		return f, cmd
	}
	
	return f, nil
}

// View renders the form
func (f TopicForm) View() string {
	var formTitle string
	if f.isEdit {
		formTitle = "Edit Topic"
	} else {
		formTitle = "Add New Topic"
	}
	
	formStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1, 2).
		Width(f.width - 4)
	
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		MarginBottom(1)
	
	inputsView := ""
	for i, input := range f.inputs {
		var label string
		switch i {
		case 0:
			label = "Topic Name:"
		case 1:
			label = "Partitions:"
		}
		
		labelStyle := lipgloss.NewStyle().Width(20)
		inputsView += labelStyle.Render(label) + " " + input.View() + "\n\n"
	}
	
	// Render buttons
	submitBgColor := "240"
	if f.buttonFocus == 0 {
		submitBgColor = "205"
	}
	
	cancelBgColor := "240"
	if f.buttonFocus == 1 {
		cancelBgColor = "205"
	}
	
	submitButtonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Background(lipgloss.Color(submitBgColor)).
		Padding(0, 3).
		MarginRight(1)
	
	cancelButtonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Background(lipgloss.Color(cancelBgColor)).
		Padding(0, 3)
	
	buttonsView := submitButtonStyle.Render(f.submitButton) + " " + cancelButtonStyle.Render(f.cancelButton)
	
	helpText := "\nUse tab/shift+tab to navigate, enter to submit"
	
	return formStyle.Render(
		titleStyle.Render(formTitle) + "\n" +
			inputsView +
			buttonsView +
			helpText,
	)
}

// TopicAddedMsg is sent when a new topic is added
type TopicAddedMsg struct {
	Name              string
	Partitions        int
	ReplicationFactor int
}

// TopicUpdatedMsg is sent when a topic is updated
type TopicUpdatedMsg struct {
	OldName    string
	Name       string
	Partitions int
}

// TopicFormCancelledMsg is sent when the topic form is cancelled
type TopicFormCancelledMsg struct{}