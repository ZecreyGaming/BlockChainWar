package config

import (
	"encoding/json"
	"os"

	"github.com/COAOX/zecrey_warrior/db"
)

type Config struct {
	Database db.Config `json:"database"`
	FPS      int       `json:"fps"`
}

func Read(configPath string) *Config {
	b, err := os.ReadFile(configPath)
	if err != nil {
		panic(err)
	}
	var config Config
	if err := json.Unmarshal(b, &config); err != nil {
		panic(err)
	}
	return &config
}
