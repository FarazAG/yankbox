package tui

import (
	"fmt"

	"github.com/FarazAG/yankstash/internal/clipboard"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	items    []clipboard.Item
	selected int
	status   string
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
		switch msg.String() {
		case "q":
			return m, tea.Quit

		case "j", "down":
			if m.selected < len(m.items)-1 {
				m.selected++
			}

		case "k", "up":
			if m.selected > 0 {
				m.selected--
			}

		case "y":
			if len(m.items) == 0 {
				break
			}

			err := clipboard.Yank(m.items[m.selected].ID)

			if err != nil {
				m.status = "Yank failed: " + err.Error()
			} else {
				m.status = "Yanked!"
			}
		}
	}

	return m, nil
}

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
		prefix := "  "

		if i == m.selected {
			prefix = "> "
		}

		view += fmt.Sprintf(
			"%s%d. %s\n",
			prefix,
			i+1,
			preview(item.Text),
		)
	}

	view += "\n" + m.status + "\n"
	view += "j/k or arrows navigate   y yank   q quit\n"

	return view
}
