# Football Wizard — Technical Specification

Interactive TUI application for predicting Brasileirão Serie A matches. Scrapes FBref for historical data, trains statistical models in pure Go, and displays results via a Bubble Tea terminal UI.

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Language | Go 1.22+ |
| CLI Visual | Bubble Tea + Lipgloss |
| Scraping | Colly |
| Database | SQLite + GORM |
| Statistics/ML | Gonum (Poisson, Logistic Regression) |
| Scheduler | robfig/cron v3 |
| Logging | Zap (Uber) |
| Config | Viper (YAML) |
| Container | Docker + Compose |

## Data Source

- **Single source**: FBref (scraped via Colly)
- Rate limiting: 1 request every 3 seconds
- Rotating User-Agents

## Prediction Models

1. **Poisson Distribution** (Gonum) — Expected goals → 1X2, O/U goals
2. **Logistic Regression** — BTTS, O/U 2.5, cards, corners
3. **Decision Tree** — Confidence level (High/Medium/Low)

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
  ├── internal/config/        →  Configuration
  ├── internal/logger/        →  Logging
  ├── internal/database/      →  Connection + migrations
  ├── internal/repository/    →  Data access
  ├── internal/scraper/       →  FBref client + parser
  ├── internal/model/         →  Features + models
  ├── internal/predictor/     →  Predictions
  ├── internal/scheduler/     →  Cron jobs
  └── internal/tui/           →  Bubble Tea (8 views)
```

## User Interface (TUI)

- Dashboard with main menu
- Scrape view (enter season → progress)
- Train view (Poisson + logistic models)
- Predict view (select teams → prediction result)
- List recent predictions
- Team stats view
- Referee profile view
- Scheduler control view

### Daemon mode

```bash
football-wizard daemon
```

Runs the scheduler headless (daily scrape at 1AM, weekly retrain Sunday 3AM).

## Roadmap

### Phase 1: Foundations ✅
- [x] Go project setup
- [x] Docker + Docker Compose
- [x] Config + Logger
- [x] SQLite + GORM + migrations
- [x] Base command with Cobra

### Phase 2: FBref Scraping ✅
- [x] HTTP client with rate limiting
- [x] HTML table parser
- [x] Match results and stats scraping

### Phase 3: Features + Poisson ✅
- [x] Feature engineering (recent form, xG)
- [x] Poisson distribution
- [x] 1X2 and O/U goals prediction

### Phase 4: Advanced ML ⚠️
- [x] Logistic regression (BTTS, cards, corners)
- [x] Feature engineering
- [ ] Decision tree with GoLearn (pending integration)

### Phase 5: CLI + Scheduler ✅
- [x] Bubble Tea with 8 views
- [x] Lipgloss styles
- [x] Cron job scheduler
- [x] Daemon mode
