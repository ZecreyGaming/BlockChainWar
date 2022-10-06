package model

import (
	"time"

	"gorm.io/gorm"
)

type Player struct {
	PlayerID  uint64 `gorm:"primaryKey" json:"player_id"`
	Name      string `json:"name"`
	Score     int    `json:"score"`
	Avatar    string `json:"avatar"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type Camp struct {
	gorm.Model
	Name      string `gorm:"uniqueIndex" json:"name"`
	ShortName string `json:"short_name"`
	Icon      string `json:"icon"`
	Score     int    `json:"score"`
}

type Game struct {
	gorm.Model
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	WinnerID  uint      `json:"winner_id"`
	Winner    Camp      `gorm:"foreignKey:WinnerID" json:"winner"`
}

type Message struct {
	gorm.Model
	Message  string `json:"message"`
	PlayerID uint64 `json:"player_id"`
	Player   Player `gorm:"foreignKey:PlayerID,references:PlayerID" json:"player"`
}
