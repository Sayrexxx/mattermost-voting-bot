package config

import (
	"strings"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

// Config represents the application configuration structure loaded from environment variables.
type Config struct {
	Mattermost struct {
		URL      string `envconfig:"MATTERMOST_URL" required:"true"`
		Token    string `envconfig:"MATTERMOST_TOKEN" required:"true"`
		TeamName string `envconfig:"MATTERMOST_TEAM" required:"true"`
		BotName  string `envconfig:"MATTERMOST_BOTNAME" default:"voting-bot"`
	}

	Tarantool struct {
		Address  string        `envconfig:"TARANTOOL_ADDRESS" default:"tarantool:3301"`
		User     string        `envconfig:"TARANTOOL_USER" default:"admin"`
		Password string        `envconfig:"TARANTOOL_PASSWORD" default:"password"`
		Timeout  time.Duration `envconfig:"TARANTOOL_TIMEOUT" default:"5s"`
		PoolSize int           `envconfig:"TARANTOOL_POOL_SIZE" default:"10"`
	}

	App struct {
		LogLevel  string `envconfig:"LOG_LEVEL" default:"info"`
		DebugMode bool   `envconfig:"DEBUG_MODE" default:"false"`
	}
}

// Load reads configuration from environment variables and returns a populated Config struct.
func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}

	cfg.Mattermost.URL = strings.TrimRight(cfg.Mattermost.URL, "/")
	return &cfg, nil
}

// SetupLogger initializes and configures a new logrus.Logger instance.
func SetupLogger(level string) *logrus.Logger {
	logger := logrus.New()
	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		logger.SetLevel(logrus.InfoLevel)
	} else {
		logger.SetLevel(lvl)
	}

	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339Nano,
	})
	return logger
}
