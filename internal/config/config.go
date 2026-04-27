package config

import (
	"os"
)

type Config struct {
	Addr   string
	DBPath string
}

func Load() *Config {
	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8080"
	}
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/gym.db"
	}
	return &Config{Addr: addr, DBPath: dbPath}
}
