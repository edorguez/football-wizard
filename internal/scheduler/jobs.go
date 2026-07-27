package scheduler

import (
	"fmt"
	"time"

	"github.com/edorguez/football-wizard/internal/predictor"
	"github.com/edorguez/football-wizard/internal/repository"
	"github.com/edorguez/football-wizard/internal/scraper"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

type Scheduler struct {
	cron           *cron.Cron
	scraper        *scraperParser
	predictor      *predictor.Engine
	log            *zap.Logger
	scrapeTime     string
	trainDay       string
	trainTime      string
	entryIDs       map[string]cron.EntryID
}

type scraperParser struct {
	client   *scraper.Client
	teams    *repository.TeamRepo
	matches  *repository.MatchRepo
	fixtures *repository.FixtureRepo
}

func NewScheduler(
	scraperClient *scraper.Client,
	teamRepo *repository.TeamRepo,
	matchRepo *repository.MatchRepo,
	fixtureRepo *repository.FixtureRepo,
	predictor *predictor.Engine,
	log *zap.Logger,
	scrapeTime, trainDay, trainTime string,
) *Scheduler {
	return &Scheduler{
		cron: cron.New(cron.WithLocation(time.Local)),
		scraper: &scraperParser{
			client:   scraperClient,
			teams:    teamRepo,
			matches:  matchRepo,
			fixtures: fixtureRepo,
		},
		predictor:  predictor,
		log:        log,
		scrapeTime: scrapeTime,
		trainDay:   trainDay,
		trainTime:  trainTime,
		entryIDs:   make(map[string]cron.EntryID),
	}
}

func (s *Scheduler) Start() error {
	s.log.Info("starting scheduler")

	scheduleDiario := fmt.Sprintf("%s %s * * *", s.scrapeTime[3:], s.scrapeTime[:2])
	id, err := s.cron.AddFunc(scheduleDiario, s.jobScrapeToday)
	if err != nil {
		return fmt.Errorf("adding daily scrape job: %w", err)
	}
	s.entryIDs["scrape_daily"] = id
	s.log.Info("scheduled daily scrape", zap.String("time", s.scrapeTime))

	trainDayNames := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
	trainDayNum := -1
	for i, day := range trainDayNames {
		if day == s.trainDay {
			trainDayNum = i
			break
		}
	}
	if trainDayNum < 0 {
		trainDayNum = 0
	}

	scheduleTrain := fmt.Sprintf("%s %s * * %d", s.trainTime[3:], s.trainTime[:2], trainDayNum)
	id, err = s.cron.AddFunc(scheduleTrain, s.jobRetrain)
	if err != nil {
		return fmt.Errorf("adding retrain job: %w", err)
	}
	s.entryIDs["retrain"] = id
	s.log.Info("scheduled weekly retrain", zap.String("day", s.trainDay), zap.String("time", s.trainTime))

	s.cron.Start()
	s.log.Info("scheduler started")
	return nil
}

func (s *Scheduler) Stop() {
	s.log.Info("stopping scheduler")
	ctx := s.cron.Stop()
	<-ctx.Done()
	s.log.Info("scheduler stopped")
}

func (s *Scheduler) IsRunning() bool {
	return len(s.cron.Entries()) > 0
}

func (s *Scheduler) jobScrapeToday() {
	s.log.Info("running daily scrape job")

	scraper := scraper.NewParser(s.scraper.client, s.log)
	fixtures, err := scraper.ParseFixtures(time.Now().Year())
	if err != nil {
		s.log.Error("error scraping fixtures", zap.Error(err))
		return
	}

	if len(fixtures) == 0 {
		s.log.Info("no upcoming fixtures found")
		return
	}

	s.log.Info("found fixtures", zap.Int("count", len(fixtures)))
}

func (s *Scheduler) jobRetrain() {
	s.log.Info("running weekly retrain job")
}
