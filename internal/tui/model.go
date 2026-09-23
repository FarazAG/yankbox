package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/FarazAG/yankstash/internal/clipboard"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type historyTickMsg struct{}

type model struct {
	items    []clipboard.Item
	selected int
	status   string
	width    int
	height   int
	viewport viewport.Model
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
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		leftWidth := m.width / 3
		rightWidth := m.width - leftWidth

		m.viewport.Width = rightWidth - 4
		m.viewport.Height = m.height - 4

		m.updateViewport()

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

			m.updateViewport()
		}

		return m, tickHistory()

	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit

		case "j", "down":
			if m.selected < len(m.items)-1 {
				m.selected++
				m.updateViewport()
			}

		case "k", "up":
			if m.selected > 0 {
				m.selected--
				m.updateViewport()
			}

		case "ctrl+d":
			m.viewport.ScrollDown(m.viewport.Height / 2)

		case "ctrl+u":
			m.viewport.ScrollUp(m.viewport.Height / 2)

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

func (m *model) updateViewport() {
	if len(m.items) == 0 {
		m.viewport.SetContent("")
		return
	}

	m.viewport.SetContent(m.items[m.selected].Text)
	m.viewport.GotoTop()
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
	if m.width == 0 || m.height == 0 {
		return ""
	}

	leftWidth := m.width / 3
	rightWidth := m.width - leftWidth

	leftStyle := lipgloss.NewStyle().
		Width(leftWidth - 2).
		Height(m.height - 2).
		Border(lipgloss.NormalBorder())

	rightStyle := lipgloss.NewStyle().
		Width(rightWidth - 2).
		Height(m.height - 2).
		Border(lipgloss.NormalBorder())

	var left strings.Builder

	fmt.Fprintln(&left, "CLIPBOARD HISTORY")
	fmt.Fprintln(&left)

	for i, item := range m.items {
		prefix := "  "

		if i == m.selected {
			prefix = "> "
		}

		fmt.Fprintf(
			&left,
			"%s%d. %s\n",
			prefix,
			i+1,
			preview(item.Text),
		)
	}

	var right strings.Builder

	fmt.Fprintln(&right, "PREVIEW")
	fmt.Fprintln(&right)
	right.WriteString(m.viewport.View())

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftStyle.Render(left.String()),
		rightStyle.Render(right.String()),
	)
}
