package main

import (
	"os"
	"time"

	"golang.org/x/exp/slog"
	"gopkg.in/yaml.v2"
)

var defaultConfig = Config{
	Global: GlobalConfig{
		ScrapeInterval: 1 * time.Minute,
		ScrapeTimeout:  10 * time.Second,
	},
	Log: LogConfig{
		Level:     slog.LevelInfo,
		Format:    "json",
		AddSource: false,
	},
	Server: ServerConfig{
		ListenAddr: ":2112",
		Endpoint:   "/metrics",
	},
}

func loadConfig(path string) (Config, error) {
	cfg := defaultConfig
	file, err := os.Open(path)
	if err != nil {
		slog.Error("Failed to open config file", slog.Any("err", err))
		os.Exit(1)
	}
	if err = yaml.NewDecoder(file).Decode(&cfg); err != nil {
		slog.Error("Failed to parse config file", slog.Any("err", err))
		os.Exit(1)
	}
	return cfg, nil
}

type Config struct {
	Global  GlobalConfig `yaml:"global"`
	Log     LogConfig    `yaml:"log"`
	Otel    OtelConfig   `yaml:"otel"`
	Server  ServerConfig `yaml:"server"`
	Configs []PlugConfig `yaml:"configs"`
}

type GlobalConfig struct {
	ScrapeInterval time.Duration `yaml:"scrape_interval"`
	ScrapeTimeout  time.Duration `yaml:"scape_timeout"`
}

type LogConfig struct {
	Level     slog.Level `yaml:"level"`
	Format    string     `yaml:"format"`
	AddSource bool       `yaml:"add_source"`
}

type OtelConfig struct {
	InstanceID string `cfg:"instance_id"`
}

type ServerConfig struct {
	ListenAddr string `yaml:"listen_addr"`
	Endpoint   string `yaml:"endpoint"`
}

type PlugConfig struct {
	Name     string        `yaml:"name"`
	Address  string        `yaml:"address"`
	Insecure bool          `yaml:"insecure"`
	Username string        `yaml:"username"`
	Password string        `yaml:"password"`
	Interval time.Duration `yaml:"interval"`
	Timeout  time.Duration `yaml:"timeout"`
}
