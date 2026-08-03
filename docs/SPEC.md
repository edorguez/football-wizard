# Football Wizard — Technical Specification

Interactive TUI application for predicting Brasileirão Serie A matches. Scrapes FBref via **HeadlessX** (self-hosted anti-detect browser platform powered by Camoufox/Firefox), trains statistical models in pure Go, and displays results through a Bubble Tea terminal UI.

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Language | Go 1.25+ |
| CLI Visual | Bubble Tea + Lipgloss |
| Scraping | HeadlessX (Camoufox/Firefox anti-detect, self-hosted via Docker) |
| Database | SQLite + GORM |
| Statistics/ML | Gonum (Poisson; Logistic Regression via gradient descent + L2) |
| Confidence | Threshold-based (probability output) |
| Scheduler | robfig/cron v3 |
| Logging | slog (stdlib) |
| Config | Viper (YAML) |
| Container | Docker + Compose (post-MVP) |
| Testing | Testify (assert, require, mock) |

## Data Source

- **Single source**: FBref, scraped via **HeadlessX**
- HeadlessX es una plataforma self-hosted que usa **Camoufox** (Firefox con parches anti-detección)
- Se despliega via Docker Compose (PostgreSQL, Redis, API, Web, Worker)
- API REST protegida con API key en `http://localhost:38473`
- Sin dependencias externas de scraping (self-hosted, sin coste por request)

## Prediction Models

1. **Poisson Distribution** (Gonum) — per-team attack/defense strength combined with league home/away scoring averages → full scoreline distribution → 1X2 and total goals O/U (configurable lines 0.5/1.5/2.5/3.5/4.5)
2. **Logistic Regression** — gradient descent + L2 on standardized causal features, one classifier per binary market:
   - BTTS, cards O/U, corners O/U, offsides O/U
   - shots O/U, shots-on-target O/U, saves O/U
   - home/away scores in 1st half, home/away scores in 2nd half
   - first scorer (home/away/neither) via two normalized classifiers
3. **Confidence Threshold** — High/Medium/Low based on probability distance from 0.5 (`|p − 0.5|`: ≥0.30 High, ≥0.15 Medium, else Low)

Auto-retrain every Sunday at 3 AM via the scheduler.

## Database

SQLite with GORM, 11 tables migrated via `AutoMigrate`:

- teams, players, team_squad_members
- matches, match_stats, match_lineups, match_player_stats, match_substitutions
- referees, fixtures, predictions

`match_stats` holds per-team full-time totals: goals, xG, shots, shots-on-target, off-target shots, saves, possession, corners, crosses, offsides, cards — plus first/second-half goals and first-goal minutes (any half and second half) that feed the half-goal and first-scorer markets.

`predictions` stores one row per saved forecast (home/away teams, 1X2 probabilities, expected goals, and the market payload as JSON) so the TUI History view survives restarts.

Planned, not yet migrated: `referee_stats`, `model_features`.

## Feature Engineering

One `Engine` (internal/model/features.go) computes **23 causal features** per fixture — only matches already played are used, so there is no look-ahead leakage and the same builder serves both training and live prediction:

- **18 per-game rates** (cumulative season, home & away): goals scored, goals conceded, xG, corners, cards, offsides, shots, shots-on-target, saves
- **4 recent-form features**: average points and average goals over the last 5 matches, home & away
- **1 Elo rating difference** (home − away)

All features are standardized before training and shared by every logistic classifier.

## Config (config.yaml)

| Key | Default | Description |
|-----|---------|-------------|
| `database.path` | `data/football-wizard.db` | SQLite file |
| `scheduler.scrape_time` / `train_day` / `train_time` | `01:00` / `Sunday` / `03:00` | daemon schedule |
| `log.level` / `log.format` | `info` / `colored` | logging |
| `headlessx.api_url` / `api_key` | `http://localhost:38473` / `""` | HeadlessX credentials (required for scraping only) |
| `model.goal_lines` | `[0.5, 1.5, 2.5, 3.5, 4.5]` | total-goals O/U lines |
| `model.over_under.cards` | `3.5` | cards O/U line |
| `model.over_under.corners` | `9.5` | corners O/U line |
| `model.over_under.offsides` | `3.5` | offsides O/U line |
| `model.over_under.shots` | `24.5` | shots O/U line |
| `model.over_under.shots_on_target` | `7.5` | shots-on-target O/U line |
| `model.over_under.saves` | `5.5` | saves O/U line |
| `model.epochs` / `learn_rate` / `l2` | `2000` / `0.05` / `0.01` | logistic training |

## Architecture

Simple layered architecture with dependency injection:

```
cmd/football-wizard/main.go  →  DI wiring + CLI flags
  ├── internal/config/        →  Viper + YAML
  ├── internal/logger/        →  slog (custom colored handler)
  ├── internal/database/      →  SQLite + GORM connection + migrations
  ├── internal/repository/    →  Data access (Team, Match, MatchStat, Fixture, Referee, Player, Lineup)
  ├── internal/scraper/
  │   ├── client.go           →  HeadlessX API client (POST /api/operators/website/scrape/html-js)
  │   ├── parser.go           →  FBref HTML → goquery → ScrapedMatch
  │   ├── match_report.go     →  match report → stats, lineups, player stats, half-goals, goal minutes
  │   ├── saver.go            →  scraped data → DB (upsert teams, refs, players, matches, stats)
  │   └── models.go           →  ScrapedMatch, ScrapedFixture, ScrapedMatchReport structs
  ├── internal/model/         →  features, poisson, logistic, market, confidence, trainer, predict, eval
  ├── internal/scheduler/     →  Cron jobs (daily scrape, weekly retrain)
  └── internal/tui/           →  Bubble Tea shell + 8 views (dashboard, scrape, train, predict,
                                 history, team, referee, scheduler) + real-time log viewer
```

The `Predictor` (prediction engine) lives in `internal/model/predict.go`.

## CLI

`football-wizard` launches the TUI by default; `daemon` runs headless; the scrape/train/predict flags remain for scripting:

```bash
football-wizard                          # launch the TUI
football-wizard ui                       # launch the TUI explicitly
football-wizard daemon                   # headless scheduler (daily scrape + weekly retrain)

# Scrape a season (requires headlessx.api_key)
football-wizard -season 2025                    # matches + fixtures
football-wizard -season 2025 -full              # + squads + match reports (backfills half-goal/first-goal data)
football-wizard -season 2025 -workers 4 -delay 3

# Train all models and report held-out accuracy
# (train/test split at round 19, per-market accuracy)
football-wizard -train

# Predict a fixture: train on all data, print 1X2 + every market
football-wizard -predict-home Flamengo -predict-away Palmeiras
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
    → POST /api/operators/website/scrape/html-js (HeadlessX API)
    → HeadlessX: Camoufox (Firefox anti-detect) navega FBref, ejecuta JS, resuelve Cloudflare
    → Extrae HTML renderizado (~800KB)
  → goquery parsea HTML en []ScrapedMatch
  → Saver.Save():
    1. Upsert unique teams → teams table
    2. Upsert unique referees → referees table
    3. Map ScrapedMatch → database.Match + database.MatchStat
    4. BulkCreate matches
  → Logs show every step in real-time

Full scrape (-full) additionally fetches each match report:
  → ParseMatchReport(html):
      - team stats (shots, shots-on-target, saves, possession, corners, cards, offsides)
      - lineups + per-player stats
      - scoring summary → first/second-half goals and goal minutes
  → Saver.SaveMatchReport() persists to match_stats, match_lineups, match_player_stats
```

Re-scraping with `-full` is required to backfill the half-goal and goal-minute columns added after the original scrape.

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
- [x] HeadlessX integration (self-hosted Docker stack + REST API)
- [x] goquery HTML parser
- [x] Saver: scraped data → SQLite (teams, refs, matches)
- [x] Real-time log viewer inside TUI

### Phase 2: Features + Poisson ✅
- [x] Feature engineering — 23 causal features (recent form, xG, Elo, per-game rates incl. shots/SOT/saves)
- [x] Poisson distribution (Gonum) — attack/defense → scoreline distribution
- [x] 1X2, total match goals O/U (0.5/1.5/2.5/3.5/4.5, configurable thresholds)

### Phase 3: Advanced ML ✅
- [x] Shared causal feature engineering (from Phase 2)
- [x] Logistic regression — BTTS, cards O/U, corners O/U, offsides O/U, shots O/U, shots-on-target O/U, saves O/U
- [x] Logistic regression — home/away scores in 1st/2nd half, first scorer
- [x] Market registry with configurable O/U thresholds (config.yaml)
- [x] Confidence threshold (High/Medium/Low)
- [x] Evaluation — train/test split accuracy report (`-train`)

### Phase 4: TUI + Scheduler ✅
- [x] Flag-based CLI (`-season`, `-full`, `-train`, `-predict-home/-predict-away`)
- [x] Bubble Tea with 8 views
- [x] Lipgloss styles
- [x] Cron job scheduler (daily scrape, weekly retrain)
- [x] Daemon mode (`football-wizard daemon`)

### Phase 5: Containerization
- [ ] Dockerfile + Docker Compose
- [ ] Chromium in Docker
- [ ] CI/CD pipeline
