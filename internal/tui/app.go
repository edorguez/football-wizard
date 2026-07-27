package tui

import (
	"github.com/edorguez/football-wizard/internal/config"
	"github.com/edorguez/football-wizard/internal/model"
	"github.com/edorguez/football-wizard/internal/predictor"
	"github.com/edorguez/football-wizard/internal/repository"
	"github.com/edorguez/football-wizard/internal/scheduler"
	"github.com/edorguez/football-wizard/internal/scraper"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AppContext struct {
	DB          *gorm.DB
	Config      *config.Config
	Log         *zap.Logger
	TeamRepo    *repository.TeamRepo
	MatchRepo   *repository.MatchRepo
	Fixtures    *repository.FixtureRepo
	Predicts    *repository.PredictionRepo
	Scraper     *scraper.Client
	Parser      *scraper.Parser
	ScrapeSaver *scraper.Saver
	Features    *model.FeatureEngine
	Trainer     *model.Trainer
	Predictor   *predictor.Engine
	Scheduler   *scheduler.Scheduler
}
