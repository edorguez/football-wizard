# Football Wizard ⚽🔮

Interactive CLI for predicting **Brasileirão Serie A** matches using historical data from **FBref**, statistical models (Poisson, Logistic Regression), and a Bubble Tea TUI.

## Features

- **Scraping** from FBref with built-in rate limiting
- **Prediction models**: Poisson (expected goals), Logistic Regression (BTTS, O/U), Decision Tree (confidence)
- **Markets**: 1X2, O/U goals, BTTS, cards, corners, first-half results
- **Automatic scheduler**: daily scrape at 1AM, weekly retrain Sunday 3AM
- **Interactive TUI** with Bubble Tea + Lipgloss
- **Daemon mode** for headless scheduler
- **Local SQLite** — no external dependencies

## Requirements

- Go 1.22+
- SQLite (bundled via CGO)

## Installation

```bash
git clone https://github.com/pc/football-wizard
cd football-wizard
make build
```

## Configuration

Edit `config.yaml`:

```yaml
database:
  path: data/football-wizard.db
scraper:
  rate_limit_seconds: 3
  user_agents:
    - "Mozilla/5.0..."
scheduler:
  scrape_time: "01:00"
  train_day: "Sunday"
  train_time: "03:00"
```

## Usage

```bash
# Start the TUI
./bin/football-wizard

# Daemon mode (scheduler in background)
./bin/football-wizard daemon

# Or with make
make run       # TUI
make daemon    # Scheduler
make dev       # go run directly
```

### Makefile commands

| Command | Description |
|---------|-------------|
| `make build` | Build binary |
| `make run` | Run TUI |
| `make daemon` | Run scheduler in background |
| `make dev` | go run directly |
| `make clean` | Clean build artifacts |
| `make docker-build` | Build Docker image |
| `make docker-run` | Docker Compose up |

## Project structure

```
football-wizard/
├── cmd/football-wizard/main.go     # Entry point
├── internal/
│   ├── config/                     # Viper + YAML config
│   ├── database/                   # SQLite + GORM + models
│   ├── repository/                 # Data access layer
│   ├── scraper/                    # Colly client + FBref parser
│   ├── model/                      # Poisson, logistic, features
│   ├── predictor/                  # Prediction orchestrator
│   ├── scheduler/                  # Cron jobs
│   ├── logger/                     # Zap logger
│   └── tui/                        # Bubble Tea (8 views)
├── pkg/utils/                      # Utilities
├── migrations/                     # Plain SQL
├── config.yaml
├── Makefile
├── Dockerfile
└── docker-compose.yml
```

## Tech stack

| Layer | Technology |
|-------|-----------|
| Language | Go 1.22+ |
| CLI Visual | Bubble Tea + Lipgloss |
| Scraping | Colly |
| Database | SQLite + GORM |
| Statistics | Gonum |
| Scheduler | robfig/cron |
| Logging | Zap |
| Config | Viper |
