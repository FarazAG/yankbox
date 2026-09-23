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
	items           []clipboard.Item
	selected        int
	status          string
	width           int
	height          int
	viewport        viewport.Model
	historyViewport viewport.Model
}

var selectedItemStyle = lipgloss.NewStyle().
	Background(lipgloss.Color("#7C3AED")).
	Foreground(lipgloss.Color("#FFFFFF")).
	Bold(true)

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
		m.viewport.Height = m.height - 6

		m.historyViewport.Width = leftWidth - 4
		m.historyViewport.Height = m.height - 6

		m.updateViewport()
		m.updateHistoryViewport()

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
			m.updateHistoryViewport()
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
				m.updateHistoryViewport()
			}

		case "k", "up":
			if m.selected > 0 {
				m.selected--
				m.updateViewport()
				m.updateHistoryViewport()
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

func (m *model) updateHistoryViewport() {
	var history strings.Builder

	for i, item := range m.items {
		prefix := "  "

		if i == m.selected {
			prefix = "▶ "
		}

		line := fmt.Sprintf(
			"%s%d. %s",
			prefix,
			i+1,
			preview(item.Text),
		)

		if i == m.selected {
			line = selectedItemStyle.Render(line)
		}

		fmt.Fprintln(&history, line)

		if i < len(m.items)-1 {
			fmt.Fprintln(&history)
		}
	}

	m.historyViewport.SetContent(history.String())

	historyItemHeight := 2
	targetOffset := m.selected * historyItemHeight

	if targetOffset < m.historyViewport.YOffset {
		m.historyViewport.SetYOffset(targetOffset)
	}

	if targetOffset >= m.historyViewport.YOffset+m.historyViewport.Height {
		m.historyViewport.SetYOffset(
			targetOffset - m.historyViewport.Height + historyItemHeight,
		)
	}
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

	left := panel(
		"YANKSTASH",
		m.historyViewport.View(),
		leftWidth,
		m.height,
	)

	right := panel(
		"PREVIEW",
		m.viewport.View(),
		rightWidth,
		m.height,
	)

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		left,
		right,
	)
}

func panel(title, content string, width, height int) string {
	borderColor := lipgloss.Color("#7C3AED")

	style := lipgloss.NewStyle().
		Width(width-2).
		Height(height-2).
		Padding(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor)

	box := style.Render(content)
	lines := strings.Split(box, "\n")

	if len(lines) == 0 {
		return box
	}

	topWidth := lipgloss.Width(lines[0])
	titleText := "── " + title + " "

	remaining := topWidth - 2 - lipgloss.Width(titleText)

	if remaining < 0 {
		remaining = 0
	}

	lines[0] = lipgloss.NewStyle().
		Foreground(borderColor).
		Bold(true).
		Render(
			"╭" +
				titleText +
				strings.Repeat("─", remaining) +
				"╮",
		)

	return strings.Join(lines, "\n")
}
