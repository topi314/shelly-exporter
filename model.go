package main

import "time"

type Config struct {
	Global  GlobalConfig `yaml:"global"`
	Server  ServerConfig `yaml:"server"`
	Configs []PlugConfig `yaml:"configs"`
}

type GlobalConfig struct {
	ScrapeInterval time.Duration `yaml:"scrape_interval"`
	ScrapeTimeout  time.Duration `yaml:"scape_timeout"`
}

type ServerConfig struct {
	ListenAddr string `yaml:"listen_addr"`
	Endpoint   string `yaml:"endpoint"`
}

type PlugConfig struct {
	Name     string        `yaml:"name"`
	Address  string        `yaml:"address"`
	Secure   bool          `yaml:"secure"`
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
	Uptime          int     `json:"uptime"`
}

type Meter struct {
	Power     float64   `json:"power"`
	Overpower float64   `json:"overpower"`
	IsValid   bool      `json:"is_valid"`
	Timestamp int       `json:"timestamp"`
	Counters  []float64 `json:"counters"`
	Total     int       `json:"total"`
}
