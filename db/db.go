package db

import (
	"time"

	"gorm.io/gorm"
)

type Player struct {
	gorm.Model
	PlayerID uint64 `json:"player_id"`
	Name     string `json:"name"`
	Score    int    `json:"score"`
	Avatar   string `json:"avatar"`
}

type Camp struct {
	gorm.Model
	Name      string `json:"name"`
	ShortName string `json:"short_name"`
	Icon      string `json:"icon"`
	Score     int    `json:"score"`
}

type Game struct {
	gorm.Model
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Winner    Camp      `json:"winner"`
}
