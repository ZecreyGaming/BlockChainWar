package model

import (
	"time"

	"gorm.io/gorm"
)

type Player struct {
	PlayerID  uint64 `gorm:"primaryKey" json:"player_id"`
	Name      string `json:"player_name"`
	Score     int    `json:"score"`
	Thumbnail string `json:"thumbnail"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type PlayerVote struct {
	GameID   uint   `gorm:"primarykey;autoIncrement:false" json:"game_id"`
	PlayerID uint64 `gorm:"primarykey;autoIncrement:false" json:"player_id"`
	Camp     uint8  `gorm:"index" json:"camp"`
}

type Camp struct {
	ID        uint8 `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	Name      string         `gorm:"uniqueIndex" json:"name"`
	ShortName string         `json:"short_name"`
	Icon      string         `json:"icon"`
	Score     int            `json:"score"`
}

type Game struct {
	gorm.Model
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	WinnerID  uint8     `json:"winner_id"`
	Winner    Camp      `gorm:"foreignKey:WinnerID" json:"winner"`
}

type Message struct {
	gorm.Model
	Message  string `json:"message"`
	PlayerID uint64 `json:"player_id"`
	Player   Player `gorm:"foreignKey:PlayerID;references:PlayerID" json:"player"`
}

const (
	Empty = iota
	BTC
	ETH
	BNB
	AVAX
	MATIC
)

var (
	BTCCamp   = Camp{ID: uint8(BTC), Name: "Bitcoin", ShortName: "BTC", Icon: "https://example.com/red.png"}
	ETHCamp   = Camp{ID: uint8(ETH), Name: "Ethereum", ShortName: "ETH", Icon: "https://example.com/blue.png"}
	BNBCamp   = Camp{ID: uint8(BNB), Name: "Binance", ShortName: "BNB", Icon: "https://example.com/green.png"}
	AVAXCamp  = Camp{ID: uint8(AVAX), Name: "Avalanche", ShortName: "AVAX", Icon: "https://example.com/yellow.png"}
	MATICCamp = Camp{ID: uint8(MATIC), Name: "Polygon", ShortName: "MATIC", Icon: "https://example.com/purple.png"}

	Camps = []Camp{
		BTCCamp,
		ETHCamp,
		BNBCamp,
		AVAXCamp,
		MATICCamp,
	}
)
