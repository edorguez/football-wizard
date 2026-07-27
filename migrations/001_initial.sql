-- Initial migration: full database schema
-- Brasileirão Serie A — Football Wizard

CREATE TABLE IF NOT EXISTS teams (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    short_name TEXT,
    city TEXT,
    founded INTEGER,
    stadium TEXT
);

CREATE TABLE IF NOT EXISTS players (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    position TEXT,
    nationality TEXT,
    birth_date TEXT,
    team_id INTEGER,
    FOREIGN KEY (team_id) REFERENCES teams(id)
);

CREATE TABLE IF NOT EXISTS matches (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date DATETIME NOT NULL,
    season INTEGER NOT NULL,
    matchday INTEGER,
    home_team_id INTEGER NOT NULL,
    away_team_id INTEGER NOT NULL,
    home_goals INTEGER,
    away_goals INTEGER,
    referee_id INTEGER,
    stadium TEXT,
    attendance INTEGER,
    FOREIGN KEY (home_team_id) REFERENCES teams(id),
    FOREIGN KEY (away_team_id) REFERENCES teams(id),
    FOREIGN KEY (referee_id) REFERENCES referees(id)
);

CREATE TABLE IF NOT EXISTS match_stats (
    match_id INTEGER PRIMARY KEY,
    home_possession REAL,
    away_possession REAL,
    home_shots INTEGER,
    away_shots INTEGER,
    home_shots_on_target INTEGER,
    away_shots_on_target INTEGER,
    home_corners INTEGER,
    away_corners INTEGER,
    home_yellow_cards INTEGER,
    away_yellow_cards INTEGER,
    home_red_cards INTEGER,
    away_red_cards INTEGER,
    home_offsides INTEGER,
    away_offsides INTEGER,
    home_fouls INTEGER,
    away_fouls INTEGER,
    home_xg REAL,
    away_xg REAL,
    FOREIGN KEY (match_id) REFERENCES matches(id)
);

CREATE TABLE IF NOT EXISTS lineups (
    match_id INTEGER,
    player_id INTEGER,
    team_id INTEGER,
    is_starter INTEGER,
    position TEXT,
    minutes_played INTEGER,
    goals INTEGER DEFAULT 0,
    assists INTEGER DEFAULT 0,
    yellow_cards INTEGER DEFAULT 0,
    red_cards INTEGER DEFAULT 0,
    PRIMARY KEY (match_id, player_id),
    FOREIGN KEY (match_id) REFERENCES matches(id),
    FOREIGN KEY (player_id) REFERENCES players(id),
    FOREIGN KEY (team_id) REFERENCES teams(id)
);

CREATE TABLE IF NOT EXISTS referees (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    nationality TEXT,
    matches_count INTEGER DEFAULT 0
);

CREATE TABLE IF NOT EXISTS referee_stats (
    referee_id INTEGER,
    match_id INTEGER,
    yellow_cards_shown INTEGER DEFAULT 0,
    red_cards_shown INTEGER DEFAULT 0,
    penalties_awarded INTEGER DEFAULT 0,
    fouls_called INTEGER DEFAULT 0,
    PRIMARY KEY (referee_id, match_id),
    FOREIGN KEY (referee_id) REFERENCES referees(id),
    FOREIGN KEY (match_id) REFERENCES matches(id)
);

CREATE TABLE IF NOT EXISTS fixtures (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date DATETIME NOT NULL,
    season INTEGER NOT NULL,
    matchday INTEGER,
    home_team_id INTEGER NOT NULL,
    away_team_id INTEGER NOT NULL,
    referee_id INTEGER,
    status TEXT DEFAULT 'scheduled',
    FOREIGN KEY (home_team_id) REFERENCES teams(id),
    FOREIGN KEY (away_team_id) REFERENCES teams(id)
);

CREATE TABLE IF NOT EXISTS predictions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    fixture_id INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    model_version TEXT,
    home_win_prob REAL,
    draw_prob REAL,
    away_win_prob REAL,
    over_0_5_prob REAL,
    over_1_5_prob REAL,
    over_2_5_prob REAL,
    over_3_5_prob REAL,
    btts_yes_prob REAL,
    over_3_5_yellow_prob REAL,
    over_4_5_yellow_prob REAL,
    home_red_prob REAL,
    away_red_prob REAL,
    over_8_5_corners_prob REAL,
    over_9_5_corners_prob REAL,
    over_10_5_corners_prob REAL,
    home_first_half_prob REAL,
    away_first_half_prob REAL,
    confidence_level TEXT,
    confidence_score REAL,
    FOREIGN KEY (fixture_id) REFERENCES fixtures(id)
);

CREATE TABLE IF NOT EXISTS model_features (
    prediction_id INTEGER PRIMARY KEY,
    home_form_goals_scored REAL,
    home_form_goals_conceded REAL,
    away_form_goals_scored REAL,
    away_form_goals_conceded REAL,
    home_xg_last5 REAL,
    away_xg_last5 REAL,
    referee_yellow_avg REAL,
    referee_red_avg REAL,
    home_corners_avg REAL,
    away_corners_avg REAL,
    is_derby INTEGER,
    distance_km REAL,
    FOREIGN KEY (prediction_id) REFERENCES predictions(id)
);

CREATE INDEX IF NOT EXISTS idx_matches_season ON matches(season);
CREATE INDEX IF NOT EXISTS idx_matches_date ON matches(date);
CREATE INDEX IF NOT EXISTS idx_matches_home_team ON matches(home_team_id);
CREATE INDEX IF NOT EXISTS idx_matches_away_team ON matches(away_team_id);
CREATE INDEX IF NOT EXISTS idx_fixtures_date ON fixtures(date);
CREATE INDEX IF NOT EXISTS idx_fixtures_status ON fixtures(status);
CREATE INDEX IF NOT EXISTS idx_predictions_created ON predictions(created_at);
CREATE INDEX IF NOT EXISTS idx_players_team ON players(team_id);
