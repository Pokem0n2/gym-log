package config

import "os"

type Config struct {
	Addr        string
	UserDataDir string
}

func Load() *Config {
	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":1118"
	}
	userDataDir := os.Getenv("USER_DATA_DIR")
	if userDataDir == "" {
		userDataDir = "./data"
	}
	return &Config{Addr: addr, UserDataDir: userDataDir}
}
