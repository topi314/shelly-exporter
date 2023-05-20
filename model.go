package main

import (
	"time"

	"golang.org/x/exp/slog"
)

type Config struct {
	Global  GlobalConfig `yaml:"global"`
	Log     LogConfig    `yaml:"log"`
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

type PlugStatus struct {
	Serial          int     `json:"serial"`
	Meters          []Meter `json:"meters"`
	Temperature     float64 `json:"temperature"`
	Overtemperature bool    `json:"overtemperature"`
	Update          Update  `json:"update"`
	Uptime          int     `json:"uptime"`
}

type Meter struct {
	Power     float64 `json:"power"`
	IsValid   bool    `json:"is_valid"`
	Overpower float64 `json:"overpower"`
	Total     int     `json:"total"`
}

type Update struct {
	HasUpdate bool `json:"has_update"`
}
