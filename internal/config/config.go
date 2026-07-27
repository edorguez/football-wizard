package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	Database  DatabaseConfig  `mapstructure:"database"`
	Scraper   ScraperConfig   `mapstructure:"scraper"`
	Scheduler SchedulerConfig `mapstructure:"scheduler"`
	Model     ModelConfig     `mapstructure:"model"`
	Log       LogConfig       `mapstructure:"log"`
}

type DatabaseConfig struct {
	Path string `mapstructure:"path"`
}

type ScraperConfig struct {
	HeadlessXURL string `mapstructure:"headlessx_url"`
	APIKey       string `mapstructure:"api_key"`
}

type SchedulerConfig struct {
	ScrapeTime string `mapstructure:"scrape_time"`
	TrainDay   string `mapstructure:"train_day"`
	TrainTime  string `mapstructure:"train_time"`
}

type ModelConfig struct {
	MinMatchesForPrediction int `mapstructure:"min_matches_for_prediction"`
	FormMatchesCount        int `mapstructure:"form_matches_count"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	configPaths := []string{
		".",
		filepath.Join(os.Getenv("HOME"), ".config", "football-wizard"),
		"/etc/football-wizard",
	}

	for _, p := range configPaths {
		v.AddConfigPath(p)
	}

	v.SetDefault("database.path", "data/football-wizard.db")
	v.SetDefault("scraper.headlessx_url", "http://localhost:38473")
	v.SetDefault("scraper.api_key", "dev-key")
	v.SetDefault("scheduler.scrape_time", "01:00")
	v.SetDefault("scheduler.train_day", "Sunday")
	v.SetDefault("scheduler.train_time", "03:00")
	v.SetDefault("model.min_matches_for_prediction", 5)
	v.SetDefault("model.form_matches_count", 5)
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "console")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, fmt.Errorf("config.yaml not found: %w", err)
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshalling config: %w", err)
	}

	return &cfg, nil
}
