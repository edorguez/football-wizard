package tui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/edorguez/football-wizard/internal/scraper"
)

type ScrapeModel struct {
	app         *AppContext
	seasonInput textinput.Model
	spinner     spinner.Model
	scraping    bool
	done        bool
	message     string
	logs        []string
	resultCh    chan error
}

func NewScrapeModel(app *AppContext) *ScrapeModel {
	ti := textinput.New()
	ti.Placeholder = "2025"
	ti.CharLimit = 4
	ti.Width = 10
	ti.Focus()

	s := spinner.New()
	s.Style = focusedStyle
	s.Spinner = spinner.Dot

	return &ScrapeModel{
		app:         app,
		seasonInput: ti,
		spinner:     s,
		logs:        []string{},
	}
}

func (m *ScrapeModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m *ScrapeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, func() tea.Msg { return switchViewMsg(dashboardView) }
		case "enter":
			if !m.scraping {
				season, err := strconv.Atoi(m.seasonInput.Value())
				if err != nil {
					m.message = errorStyle.Render("invalid season")
					return m, nil
				}
				m.scraping = true
				m.done = false
				m.message = ""
				m.logs = []string{}
				m.resultCh = make(chan error, 1)
				go m.doScrape(season, m.resultCh)
				return m, tea.Batch(m.spinner.Tick, m.listenProgress())
			}
		}
	case progressLogMsg:
		m.logs = append(m.logs, string(msg))
		if len(m.logs) > 100 {
			m.logs = m.logs[len(m.logs)-100:]
		}
		if m.scraping {
			return m, m.listenProgress()
		}
		return m, nil
	case progressHeartbeat:
		if m.scraping {
			return m, m.listenProgress()
		}
		return m, nil
	case scrapeResultMsg:
		m.scraping = false
		result := string(msg)
		if strings.HasPrefix(result, "OK:") {
			m.done = true
			m.message = successStyle.Render(fmt.Sprintf("scraping complete — %s", result[3:]))
		} else {
			m.message = errorStyle.Render(fmt.Sprintf("error: %s", result))
		}
		return m, nil
	case spinner.TickMsg:
		if m.scraping {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	}

	var cmd tea.Cmd
	m.seasonInput, cmd = m.seasonInput.Update(msg)
	return m, cmd
}

func (m *ScrapeModel) View() string {
	s := titleStyle.Render("📊 Scrape Data")
	s += "\n\n"

	if !m.scraping && !m.done && m.message == "" {
		s += headerStyle.Render("season to scrape:")
		s += "\n\n"
		s += m.seasonInput.View()
		s += "\n\n"
		s += helpStyle.Render("Enter to start • Esc to go back")
	}

	if m.scraping || len(m.logs) > 0 {
		s += logBoxStyle.Render(m.renderLogs())
		s += "\n"
	}

	if m.scraping {
		s += m.spinner.View() + " scraping in progress\n"
	}

	if !m.scraping && m.message != "" {
		s += "\n" + m.message + "\n"
	}

	if !m.scraping && (m.done || m.message != "") {
		s += "\n" + helpStyle.Render("Esc to go back to menu")
	}

	return s
}

func (m *ScrapeModel) renderLogs() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(" %s\n", headerStyle.Render("Scraping Log")))
	b.WriteString(fmt.Sprintf(" %s\n", dimmedStyle.Render(strings.Repeat("─", 50))))

	if len(m.logs) == 0 {
		b.WriteString(fmt.Sprintf(" %s\n", dimmedStyle.Render("waiting...")))
	} else {
		start := 0
		if len(m.logs) > 20 {
			start = len(m.logs) - 20
		}
		for _, entry := range m.logs[start:] {
			ts := time.Now().Format("15:04:05")
			line := fmt.Sprintf(" %s %s %s",
				logTimeStyle.Render(ts),
				logPrefixStyle.Render("▸"),
				logTextStyle.Render(entry),
			)
			b.WriteString(line + "\n")
		}
	}

	return b.String()
}

func (m *ScrapeModel) listenProgress() tea.Cmd {
	return func() tea.Msg {
		select {
		case msg := <-m.app.Scraper.Progress:
			return progressLogMsg(msg.Text)
		case err := <-m.resultCh:
			if err != nil {
				return scrapeResultMsg(fmt.Sprintf("ERR:%s", err))
			}
			return scrapeResultMsg("OK:season completed")
		case <-time.After(500 * time.Millisecond):
			return progressHeartbeat("")
		}
	}
}

func (m *ScrapeModel) doScrape(season int, done chan error) {
	scraped, err := m.app.Parser.ParseMatchResults(season)
	if err != nil {
		done <- err
		return
	}

	m.app.Scraper.Progress <- scraper.ProgressMsg{
		Text: fmt.Sprintf("saving %d matches to database...", len(scraped)),
	}

	saved, err := m.app.ScrapeSaver.Save(scraped)
	if err != nil {
		done <- err
		return
	}

	m.app.Scraper.Progress <- scraper.ProgressMsg{
		Text: fmt.Sprintf("saved %d matches to database", saved),
	}
	done <- nil
}

type progressLogMsg string
type progressHeartbeat string
type scrapeResultMsg string
