package tui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/edorguez/football-wizard/internal/logger"
)

const scrapeHeaderHeight = 6

type ScrapeModel struct {
	deps     Deps
	input    textinput.Model
	viewport viewport.Model
	spinner  spinner.Model
	buffer   *logger.RingBuffer
	season   int
	full     bool
	running  bool
	done     bool
	err      error
	ready    bool
}

func NewScrapeModel(deps Deps) tea.Model {
	input := textinput.New()
	input.Placeholder = "e.g. 2025"
	input.CharLimit = 4
	input.Focus()

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = SpinnerStyle

	return &ScrapeModel{deps: deps, input: input, spinner: s}
}

func (m *ScrapeModel) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.spinner.Tick)
}

type scrapeStartMsg struct {
	season int
	full   bool
}

type scrapeDoneMsg struct {
	err error
}

type scrapeTickMsg time.Time

func (m *ScrapeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if !m.ready {
			m.viewport = viewport.New(msg.Width-4, msg.Height-scrapeHeaderHeight)
			m.ready = true
		} else {
			m.viewport.Width = msg.Width - 4
			m.viewport.Height = msg.Height - scrapeHeaderHeight
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.running {
				return m, nil
			}
			return m, backCmd()
		case "ctrl+c":
			return m, tea.Quit
		}

		if m.running {
			var cmd tea.Cmd
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		}

		if msg.Type == tea.KeyEnter {
			return m, m.start()
		}
		if msg.String() == "f" {
			m.full = !m.full
			return m, nil
		}
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd

	case scrapeStartMsg:
		m.running = true
		m.done = false
		m.err = nil
		m.buffer = logger.NewRingBuffer(500)
		m.viewport.SetContent("")
		m.viewport.GotoTop()

		scrapeLog := logger.New(logger.Config{
			Level:  m.deps.Cfg.Log.Level,
			Format: "text",
			Output: m.buffer,
		})
		sc := m.deps.NewScraper(scrapeLog)

		run := func() tea.Msg {
			var err error
			if msg.full {
				err = sc.ScrapeSeasonFull(msg.season)
			} else {
				err = sc.ScrapeSeason(msg.season)
			}
			return scrapeDoneMsg{err: err}
		}
		return m, tea.Batch(run, m.tick(), m.spinner.Tick)

	case scrapeDoneMsg:
		m.running = false
		m.done = true
		m.err = msg.err
		m.refreshLogs()
		return m, nil

	case scrapeTickMsg:
		m.refreshLogs()
		return m, m.tick()

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		if !m.running {
			return m, nil
		}
		return m, cmd
	}

	return m, nil
}

func (m *ScrapeModel) start() tea.Cmd {
	season, err := strconv.Atoi(strings.TrimSpace(m.input.Value()))
	if err != nil || season < 1990 || season > 2100 {
		m.err = fmt.Errorf("enter a valid season year (e.g. 2025)")
		return nil
	}
	m.season = season
	m.done = false
	return func() tea.Msg {
		return scrapeStartMsg{season: season, full: m.full}
	}
}

func (m *ScrapeModel) tick() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return scrapeTickMsg(t)
	})
}

func (m *ScrapeModel) refreshLogs() {
	if m.buffer == nil || !m.ready {
		return
	}
	m.viewport.SetContent(strings.Join(m.buffer.Lines(), "\n"))
	if m.running {
		m.viewport.GotoBottom()
	}
}

func (m *ScrapeModel) View() string {
	var b strings.Builder
	b.WriteString(TitleStyle("Scrape a Season"))
	b.WriteString("\n")

	if !m.ready {
		b.WriteString("\nloading...")
		return b.String()
	}

	fmt.Fprintf(&b, "Season: %s\n", m.input.View())

	full := "no"
	if m.full {
		full = "yes"
	}
	fmt.Fprintf(&b, "Full scrape (squads + match reports): %s  (press f to toggle)\n", full)

	switch {
	case m.running:
		b.WriteString(WarningStyle(fmt.Sprintf("%s scraping season %d...", m.spinner.View(), m.season)))
	case m.done && m.err != nil:
		b.WriteString(ErrorStyle(fmt.Sprintf("scrape failed: %v", m.err)))
	case m.done:
		b.WriteString(SuccessStyle(fmt.Sprintf("season %d scraped", m.season)))
	case m.err != nil:
		b.WriteString(ErrorStyle(m.err.Error()))
	default:
		b.WriteString(MutedStyle("idle — enter a season and press enter"))
	}
	b.WriteString("\n\n")

	b.WriteString(m.viewport.View())
	b.WriteString("\n")
	b.WriteString(HelpStyle("enter: start · f: toggle full · pgup/pgdn: scroll · esc: menu"))
	return b.String()
}
