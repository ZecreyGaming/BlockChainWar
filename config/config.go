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
	AccountName       string    `json:"account_name"`
	Seed              string    `json:"seed"`
	NftPrefix         string    `json:"nft_prefix"`
	CollectionId      int64     `json:"collection_id"`
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
