# Football Wizard ⚽🔮

Interactive TUI for predicting **Brasileirão Serie A** matches. Scrapes FBref via **go-rod + go-rod/stealth** (Chrome headless with anti-detection), trains statistical models in pure Go, and displays results through a Bubble Tea terminal UI.

## Features

- **Scraping** from FBref via go-rod + stealth — Chrome headless que pasa Cloudflare
- **Real-time log viewer** inside the TUI showing every scraping step
- **Auto-save** scraped data to local SQLite
- **Prediction models**: Poisson (expected goals), Logistic Regression (BTTS, O/U), Decision Tree (confidence)
- **Markets**: 1X2, O/U goals, BTTS, cards, corners, first-half results
- **Automatic scheduler**: daily scrape at 1AM, weekly retrain Sunday 3AM
- **Interactive TUI** with Bubble Tea + Lipgloss
- **Daemon mode** for headless scheduler

## Requirements

- Go 1.25+
- Chrome or Chromium (for go-rod browser engine)
- SQLite (bundled via CGO)

## Quick start

```bash
# 1. Install Chrome (macOS)
brew install --cask google-chrome

# Or Chromium (Linux)
sudo apt install chromium-browser

# 2. Clone and build
git clone https://github.com/edorguez/football-wizard
cd football-wizard
make build

# 3. Run the app
make run
```

## Configuration

Edit `config.yaml`:

```yaml
database:
  path: data/football-wizard.db
scheduler:
  scrape_time: "01:00"
  train_day: "Sunday"
  train_time: "03:00"
```

## Usage

```bash
./bin/football-wizard

# Or with make
make run       # TUI
make daemon    # Scheduler headless
make dev       # go run directly
```

### Makefile commands

| Command | Description |
|---------|-------------|
| `make build` | Build binary |
| `make run` | Run TUI |
| `make daemon` | Run scheduler headless |
| `make dev` | go run directly |
| `make test` | Run all tests |
| `make clean` | Clean build artifacts |

## Project structure

```
football-wizard/
├── cmd/football-wizard/main.go     # Entry point + DI wiring
├── internal/
│   ├── config/                     # Viper + YAML config
│   ├── database/                   # SQLite + GORM + models
│   ├── repository/                 # Data access layer (Team, Match, Fixture)
│   ├── scraper/
│   │   ├── client.go              # go-rod + stealth browser client
│   │   ├── parser.go              # FBref HTML parser (goquery)
│   │   ├── saver.go               # Saves scraped data to database
│   │   └── models.go              # Scraped data structs
│   ├── model/                      # Poisson, logistic, confidence, trainer
│   ├── predictor/                  # Prediction orchestrator
│   ├── scheduler/                  # Cron jobs (daily scrape, weekly retrain)
│   ├── logger/                     # slog logger
│   └── tui/                        # Bubble Tea (8 views + real-time logs)
├── config.yaml
├── Makefile
└── README.md
```

## Tech stack

| Layer | Technology |
|-------|-----------|
| Language | Go 1.25+ |
| CLI Visual | Bubble Tea + Lipgloss |
| Scraping | go-rod + go-rod/stealth |
| Database | SQLite + GORM |
| Statistics/ML | Gonum (Poisson, Logistic Regression) |
| Confidence | Threshold-based |
| Scheduler | robfig/cron |
| Logging | slog (stdlib) |
| Config | Viper |
| Testing | Testify |
