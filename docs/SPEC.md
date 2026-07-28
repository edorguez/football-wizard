# Football Wizard — Technical Specification

Interactive TUI application for predicting Brasileirão Serie A matches. Scrapes FBref via **go-rod + go-rod/stealth** (Chrome headless with anti-detection), trains statistical models in pure Go, and displays results through a Bubble Tea terminal UI.

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Language | Go 1.25+ |
| CLI Visual | Bubble Tea + Lipgloss |
| Scraping | go-rod + go-rod/stealth (Chrome headless) |
| Database | SQLite + GORM |
| Statistics/ML | Gonum (Poisson, Logistic Regression) |
| Confidence | Threshold-based (probability output) |
| Scheduler | robfig/cron v3 |
| Logging | slog (stdlib) |
| Config | Viper (YAML) |
| Container | Docker + Compose (post-MVP) |
| Testing | Testify (assert, require, mock) |

## Data Source

- **Single source**: FBref, scraped via **go-rod + go-rod/stealth**
- go-rod lanza Chrome headless real que ejecuta JavaScript
- go-rod/stealth module parchea señales de detección:
  - `navigator.webdriver` (ocultado)
  - `navigator.plugins`, `languages`, `hardwareConcurrency`
  - WebGL fingerprint (canvas renderer)
  - `navigator.permissions`
- Sin dependencias externas (no requiere APIs de terceros)

## Prediction Models

1. **Poisson Distribution** (Gonum) — Expected goals → 1X2, O/U goals
2. **Logistic Regression** — BTTS, O/U 2.5, cards, corners
3. **Confidence Threshold** — High/Medium/Low based on probability spread from Poisson & Logistic outputs

Auto-retrain every Sunday at 3 AM.

## Database

SQLite with GORM. 10 tables:

- teams, players, matches, match_stats, lineups
- referees, referee_stats, fixtures
- predictions, model_features

## Architecture

Simple layered architecture with dependency injection:

```
cmd/football-wizard/main.go  →  DI wiring
  ├── internal/config/        →  Viper + YAML
  ├── internal/logger/        →  Zap
  ├── internal/database/      →  SQLite + GORM connection + migrations
  ├── internal/repository/    →  Data access (Team, Match, Fixture, Prediction, Referee)
  ├── internal/scraper/
  │   ├── client.go           →  go-rod + stealth browser client
  │   ├── parser.go           →  FBref HTML → goquery → ScrapedMatch
  │   ├── saver.go            →  ScrapedMatch → DB (upsert teams, refs, matches)
  │   └── models.go           →  ScrapedMatch, ScrapedFixture structs
  ├── internal/model/         →  Poisson, Logistic, Features, Confidence, Trainer
  ├── internal/predictor/     →  Prediction engine
  ├── internal/scheduler/     →  Cron jobs
  └── internal/tui/           →  8 Bubble Tea views + real-time log viewer
```

## User Interface (TUI)

- **Dashboard** — main menu with 7 options
- **Scrape view** — enter a season, see real-time scraping logs, auto-saved to DB
- **Train view** — train Poisson + logistic models
- **Predict view** — select home/away teams → see probabilities
- **List recent predictions** — browse prediction history
- **Team stats view** — search a team, see recent form
- **Referee profile view** — search a referee
- **Scheduler control view** — start/stop the automatic scheduler

### Scraping pipeline

```
User enters season → Parser.ParseMatchResults(season)
  → Client.FetchHTML(url):
    → go-rod: lanza Chrome headless
    → go-rod/stealth: parchea detección (webdriver, plugins, WebGL)
    → Chrome navega FBref, ejecuta JS, resuelve Cloudflare
    → Extrae HTML renderizado (~800KB)
  → goquery parsea HTML en []ScrapedMatch
  → Saver.Save():
    1. Upsert unique teams → teams table
    2. Upsert unique referees → referees table
    3. Map ScrapedMatch → database.Match + database.MatchStat
    4. BulkCreate matches
  → Logs show every step in real-time
```

### Daemon mode

```bash
football-wizard daemon
```

Runs the scheduler headless (daily scrape at 1AM, weekly retrain Sunday 3AM).

## Roadmap

### Phase 1: Foundations + Scraping
- [x] Go project setup (module: github.com/edorguez/football-wizard)
- [x] Config (Viper) + Logger (slog)
- [x] SQLite + GORM + migrations (teams, refs, matches, match_stats, fixtures)
- [x] Makefile with dev commands
- [x] go-rod + stealth integration
- [x] goquery HTML parser
- [x] Saver: scraped data → SQLite (teams, refs, matches)
- [ ] Real-time log viewer inside TUI

### Phase 2: Features + Poisson ✅
- [ ] Feature engineering (recent form, xG)
- [ ] Poisson distribution (Gonum)
- [ ] 1X2 and O/U goals prediction

### Phase 3: Advanced ML ⚠️
- [ ] Logistic regression (BTTS, cards, corners)
- [ ] Feature engineering
- [ ] Confidence threshold (probability-based)

### Phase 4: CLI + Scheduler ✅
- [ ] Bubble Tea with 8 views
- [ ] Lipgloss styles
- [ ] Cron job scheduler
- [ ] Daemon mode

### Phase 5: Containerization
- [ ] Dockerfile + Docker Compose
- [ ] Chromium in Docker
- [ ] CI/CD pipeline
