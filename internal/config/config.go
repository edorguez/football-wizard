package config

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Database  DatabaseConfig   `mapstructure:"database"`
	Scheduler SchedulerConfig  `mapstructure:"scheduler"`
	Log       LogConfig        `mapstructure:"log"`
	HeadlessX HeadlessXConfig  `mapstructure:"headlessx"`
}

type DatabaseConfig struct {
	Path string `mapstructure:"path"`
}

type SchedulerConfig struct {
	ScrapeTime string `mapstructure:"scrape_time"`
	TrainDay   string `mapstructure:"train_day"`
	TrainTime  string `mapstructure:"train_time"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

type HeadlessXConfig struct {
	APIURL string `mapstructure:"api_url"`
	APIKey string `mapstructure:"api_key"`
}

func Load(path string) (*Config, error) {
	if err := godotenv.Load(); err != nil {
		slog.Debug("no .env file found, using config.yaml and system env")
	}

	v := viper.New()

	v.SetConfigFile(path)
	v.SetConfigType("yaml")

	v.SetDefault("database.path", "data/football-wizard.db")
	v.SetDefault("scheduler.scrape_time", "01:00")
	v.SetDefault("scheduler.train_day", "Sunday")
	v.SetDefault("scheduler.train_time", "03:00")
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "json")
	v.SetDefault("headlessx.api_url", "http://localhost:38473")
	v.SetDefault("headlessx.api_key", "")

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}

	return &cfg, nil
}
