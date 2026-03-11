package tui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	apptracking "game-time-tracker/internal/application/tracking"
	historydomain "game-time-tracker/internal/domain/history"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	appStyle = lipgloss.NewStyle().Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39"))

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("246"))

	activeTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("33")).
			Padding(0, 1)

	inactiveTabStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("250")).
				Padding(0, 1)

	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("238")).
			Padding(1, 2)

	headerRowStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("227"))

	rowGameStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
	rowTimeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("80"))
	rowDateStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("250"))

	controlsStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	errorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("203"))

	statusTrackingStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("15")).
				Background(lipgloss.Color("35")).
				Padding(0, 1)

	statusPausedStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("232")).
				Background(lipgloss.Color("214")).
				Padding(0, 1)

	statusMonitoringStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("232")).
				Background(lipgloss.Color("250")).
				Padding(0, 1)
)

type SignalMsg struct {
	Signal string
}

type statusMsg struct {
	status apptracking.RuntimeStatus
}

type historyMsg struct {
	entries []historydomain.Entry
}

type runtimeErrMsg struct {
	err error
}

type historyRow struct {
	totalSecs  int64
	lastPlayed time.Time
}

type Model struct {
	statusCh  <-chan apptracking.RuntimeStatus
	historyCh <-chan []historydomain.Entry
	errCh     <-chan error

	viewIndex int
	status    apptracking.RuntimeStatus
	history   map[string]historyRow
	lastErr   string
}

func NewModel(
	statusCh <-chan apptracking.RuntimeStatus,
	historyCh <-chan []historydomain.Entry,
	errCh <-chan error,
	initialHistory []historydomain.Entry,
) Model {
	history := make(map[string]historyRow, len(initialHistory))
	for _, e := range initialHistory {
		row := history[e.GameName]
		row.totalSecs += e.TotalTimeSecs
		if e.LastPlayedDate.After(row.lastPlayed) {
			row.lastPlayed = e.LastPlayedDate
		}
		history[e.GameName] = row
	}

	return Model{
		statusCh:  statusCh,
		historyCh: historyCh,
		errCh:     errCh,
		viewIndex: 0,
		history:   history,
		status: apptracking.RuntimeStatus{
			State:   "monitoring",
			Updated: time.Now(),
		},
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(waitForStatus(m.statusCh), waitForHistory(m.historyCh), waitForError(m.errCh))
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "tab", "right":
			m.viewIndex = (m.viewIndex + 1) % 2
		case "left":
			m.viewIndex = (m.viewIndex + 1) % 2
		case "1":
			m.viewIndex = 0
		case "2":
			m.viewIndex = 1
		}
		return m, nil
	case SignalMsg:
		m.lastErr = fmt.Sprintf("received signal %s", msg.Signal)
		return m, tea.Quit
	case statusMsg:
		m.status = msg.status
		return m, waitForStatus(m.statusCh)
	case historyMsg:
		for _, entry := range msg.entries {
			row := m.history[entry.GameName]
			row.totalSecs += entry.TotalTimeSecs
			if entry.LastPlayedDate.After(row.lastPlayed) {
				row.lastPlayed = entry.LastPlayedDate
			}
			m.history[entry.GameName] = row
		}
		return m, waitForHistory(m.historyCh)
	case runtimeErrMsg:
		if msg.err != nil {
			m.lastErr = msg.err.Error()
		}
		return m, waitForError(m.errCh)
	default:
		return m, nil
	}
}

func (m Model) View() string {
	header := titleStyle.Render("Game Time Tracker")
	tabs := m.renderTabs()
	subtitle := subtitleStyle.Render("Track daily playtime with live status and persisted history")
	body := m.dashboardView()
	if m.viewIndex == 1 {
		body = m.statusView()
	}

	controls := controlsStyle.Render("Controls: 1 Dashboard | 2 Active Status | tab/right/left switch views | q quit")

	footer := ""
	if m.lastErr != "" {
		footer = "\n" + errorStyle.Render("Last event: "+m.lastErr)
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		subtitle,
		tabs,
		panelStyle.Render(body),
		controls,
	)

	return appStyle.Render(content) + footer + "\n"
}

func (m Model) dashboardView() string {
	if len(m.history) == 0 {
		return "Dashboard\n\nNo history yet. Launch a watched game to start collecting playtime."
	}

	games := make([]string, 0, len(m.history))
	for game := range m.history {
		games = append(games, game)
	}
	sort.Strings(games)

	const gameWidth = 30
	header := lipgloss.JoinHorizontal(
		lipgloss.Left,
		headerRowStyle.Width(gameWidth).Render("Game"),
		headerRowStyle.Width(12).Render("Total Time"),
		headerRowStyle.Render("Last Played"),
	)

	lines := []string{"Dashboard", "", header}
	for _, game := range games {
		row := m.history[game]
		lastPlayed := "-"
		if !row.lastPlayed.IsZero() {
			lastPlayed = row.lastPlayed.Format("2006-01-02 15:04:05")
		}

		line := lipgloss.JoinHorizontal(
			lipgloss.Left,
			rowGameStyle.Width(gameWidth).Render(game),
			rowTimeStyle.Width(12).Render(formatSeconds(row.totalSecs)),
			rowDateStyle.Render(lastPlayed),
		)
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func (m Model) statusView() string {
	lines := []string{"Active Status", ""}
	switch m.status.State {
	case "tracking":
		lines = append(lines,
			"State: "+statusTrackingStyle.Render("TRACKING"),
			"Game:  "+m.status.GameName,
			"Today: "+formatDuration(m.status.Elapsed),
		)
	case "paused":
		lines = append(lines,
			"State: "+statusPausedStyle.Render("PAUSED"),
			"Game:  "+m.status.GameName,
			"Today: "+formatDuration(m.status.Elapsed),
		)
	default:
		lines = append(lines,
			"State: "+statusMonitoringStyle.Render("MONITORING"),
			"Game:  -",
			"Today: 00:00:00",
		)
	}
	lines = append(lines, "Updated: "+m.status.Updated.Format("15:04:05"))
	return strings.Join(lines, "\n")
}

func (m Model) renderTabs() string {
	dashboard := inactiveTabStyle.Render("1 Dashboard")
	status := inactiveTabStyle.Render("2 Active Status")
	if m.viewIndex == 0 {
		dashboard = activeTabStyle.Render("1 Dashboard")
	} else {
		status = activeTabStyle.Render("2 Active Status")
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, dashboard, " ", status)
}

func waitForStatus(ch <-chan apptracking.RuntimeStatus) tea.Cmd {
	return func() tea.Msg {
		status, ok := <-ch
		if !ok {
			return runtimeErrMsg{}
		}
		return statusMsg{status: status}
	}
}

func waitForHistory(ch <-chan []historydomain.Entry) tea.Cmd {
	return func() tea.Msg {
		entries, ok := <-ch
		if !ok {
			return runtimeErrMsg{}
		}
		return historyMsg{entries: entries}
	}
}

func waitForError(ch <-chan error) tea.Cmd {
	return func() tea.Msg {
		err, ok := <-ch
		if !ok {
			return runtimeErrMsg{}
		}
		return runtimeErrMsg{err: err}
	}
}

func formatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

func formatSeconds(total int64) string {
	if total < 0 {
		total = 0
	}
	h := total / 3600
	m := (total % 3600) / 60
	s := total % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}
