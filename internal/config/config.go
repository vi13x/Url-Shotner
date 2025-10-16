package config

import (
	"flag"
	"time"
)

type Config struct {
	Addr            string
	BaseURL         string
	ShutdownTimeout time.Duration
	RateLimit       int
	RateWindow      time.Duration
}

func Load() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.Addr, "addr", ":8080", "Server address")
	flag.StringVar(&cfg.BaseURL, "base-url", "http://localhost:8080", "Base URL for short links")
	flag.DurationVar(&cfg.ShutdownTimeout, "shutdown-timeout", 10*time.Second, "Graceful shutdown timeout")
	flag.IntVar(&cfg.RateLimit, "rate-limit", 100, "Rate limit per window")
	flag.DurationVar(&cfg.RateWindow, "rate-window", time.Minute, "Rate limit window")

	flag.Parse()
	return cfg
}
