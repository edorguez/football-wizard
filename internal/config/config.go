package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Database  DatabaseConfig  `mapstructure:"database"`
	Scheduler SchedulerConfig `mapstructure:"scheduler"`
	Log       LogConfig       `mapstructure:"log"`
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

func Load(path string) (*Config, error) {
	v := viper.New()

	v.SetConfigFile(path)
	v.SetConfigType("yaml")

	v.SetDefault("database.path", "data/football-wizard.db")
	v.SetDefault("scheduler.scrape_time", "01:00")
	v.SetDefault("scheduler.train_day", "Sunday")
	v.SetDefault("scheduler.train_time", "03:00")
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "json")

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
