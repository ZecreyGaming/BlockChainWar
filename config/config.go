package config

import (
	"encoding/json"
	"os"

	"github.com/COAOX/zecrey_warrior/db"
)

const (
	ChatRoomName = "chat"
	GameRoomName = "game"
)

type Config struct {
	Database          db.Config `json:"database"`
	FPS               int       `json:"fps"`
	GameRoundInterval int       `json:"game_round_interval"`
	FrontendType      string    `json:"frontend_type"`
	ItemFrameChance   int       `json:"item_frame_chance"`
	GameDuration      int       `json:"game_duration"`
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
