package tui

import (
	"fmt"
	"time"

	"github.com/FarazAG/yankstash/internal/clipboard"
	tea "github.com/charmbracelet/bubbletea"
)

type historyTickMsg struct{}

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
	return tickHistory()
}

func tickHistory() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(time.Time) tea.Msg {
		return historyTickMsg{}
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case historyTickMsg:
		items, err := clipboard.History()
		if err != nil {
			m.status = "Refresh failed: " + err.Error()
			return m, tickHistory()
		}

		if historyChanged(m.items, items) {
			m.items = items

			if m.selected >= len(m.items) {
				m.selected = len(m.items) - 1
			}

			if m.selected < 0 {
				m.selected = 0
			}
		}

		return m, tickHistory()

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

		case "1", "2", "3", "4", "5", "6", "7", "8", "9", "0":
			index := int(msg.Runes[0] - '1')

			if msg.String() == "0" {
				index = 9
			}

			if index >= len(m.items) {
				break
			}

			err := clipboard.Yank(m.items[index].ID)

			if err != nil {
				m.status = "Yank failed: " + err.Error()
			} else {
				m.status = fmt.Sprintf("Yanked #%d", index+1)
			}
		}
	}

	return m, nil
}

func historyChanged(oldItems, newItems []clipboard.Item) bool {
	if len(oldItems) != len(newItems) {
		return true
	}

	for i := range oldItems {
		if oldItems[i].ID != newItems[i].ID {
			return true
		}
	}

	return false
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
