# Football Wizard ⚽🔮

Interactive TUI for predicting **Brasileirão Serie A** matches. Scrapes FBref via **HeadlessX** (self-hosted anti-detect browser platform), trains statistical models in pure Go, and displays results through a Bubble Tea terminal UI.

## Features

- **Scraping** from FBref via HeadlessX — self-hosted Camoufox/Firefox que pasa Cloudflare
- **Real-time log viewer** inside the TUI showing every scraping step
- **Auto-save** scraped data to local SQLite
- **Prediction models**: Poisson (expected goals), Logistic Regression (BTTS, O/U), Decision Tree (confidence)
- **Markets**: 1X2, O/U goals, BTTS, cards, corners, first-half results
- **Automatic scheduler**: daily scrape at 1AM, weekly retrain Sunday 3AM
- **Interactive TUI** with Bubble Tea + Lipgloss
- **Daemon mode** for headless scheduler

## Requirements

- Go 1.25+
- Docker with Compose v2 (for HeadlessX)
- Node.js 22+ (for HeadlessX CLI)
- SQLite (bundled via CGO)

## Quick start

```bash
# 1. Install HeadlessX (self-hosted scraping platform)
npm install -g @headlessx-cli/core
headlessx init --mode self-host
headlessx status

# 2. Create an API key from http://localhost:34872 or via the API

# 3. Clone and build
git clone https://github.com/edorguez/football-wizard
cd football-wizard
make build

# 4. Set your HeadlessX API key
# Edit config.yaml or set HEADLESSX_API_KEY env var

# 5. Run the app
make dev -- --season 2025
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
headlessx:
  api_url: http://localhost:38473
  api_key: ""  # required - set your HeadlessX API key here
```

All values can be overridden via environment variables (e.g. `DATABASE_PATH`, `LOG_LEVEL`). Copy `.env.example` to `.env` for local overrides.

## HeadlessX (Scraping Service)

This project uses [HeadlessX](https://github.com/saifyxpro/HeadlessX) to scrape FBref. HeadlessX is a self-hosted anti-detect browser platform that bypasses Cloudflare using Camoufox (Firefox). It runs in a Docker stack:

- Dashboard: http://localhost:34872
- API: http://localhost:38473

**First-time setup:**
```bash
npm install -g @headlessx-cli/core
headlessx init --mode self-host
headlessx status
```

Then create an API key from the dashboard at `http://localhost:34872` and add it to `config.yaml` or the `HEADLESSX_API_KEY` env var.

## Data Storage

Scraped match data is stored in a local SQLite database at `data/football-wizard.db` by default. The `data/` directory is created automatically by `make build`, `make dev`, and `make run`.

- **Database location**: configurable via `database.path` in `config.yaml` or `DATABASE_PATH` env var
- **Data directory**: listed in `.gitignore` — your scraped data stays local
- **Reset data**: `make clean` removes both `bin/` and `data/`

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
│   │   ├── client.go              # HeadlessX API client
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
| Scraping | HeadlessX (Camoufox/Firefox anti-detect) |
| Database | SQLite + GORM |
| Statistics/ML | Gonum (Poisson, Logistic Regression) |
| Confidence | Threshold-based |
| Scheduler | robfig/cron |
| Logging | slog (stdlib) |
| Config | Viper |
| Testing | Testify |
