package tui

import (
	"fmt"

	"github.com/FarazAG/yankstash/internal/clipboard"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	items []clipboard.Item
}

func NewModel() model {
	items, err := clipboard.History()
	if err != nil {
		panic(err)
	}

	return model{
		items: items,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" {
			return m, tea.Quit
		}
	}
	return m, nil
}

// view helper function - preview

func preview(text string) string {
	for i, char := range text {
		if char == '\n' || char == '\r' {
			return text[:i] + "..."
		}
	}

	return text
}

func (m model) View() string {
	var view string

	for i, item := range m.items {
		view += fmt.Sprintf("%d. %s\n", i+1, preview(item.Text))
	}

	view += "\nPress q to quit.\n"

	return view
}
