package db

import (
	"github.com/COAOX/zecrey_warrior/model"
	"gorm.io/gorm"
)

type Config struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
}

type Client struct {
	*gorm.DB
	Game    game
	Camp    camp
	Player  player
	Message message
}

type db struct {
	*gorm.DB
}

// BTC
// ETH
// BNB
// AVAX
// MATIC
var Camps = []model.Camp{
	{Name: "Bitcoin", ShortName: "BTC", Icon: "https://example.com/red.png"},
	{Name: "Ethereum", ShortName: "ETH", Icon: "https://example.com/blue.png"},
	{Name: "Binance", ShortName: "BNB", Icon: "https://example.com/green.png"},
	{Name: "Avalanche", ShortName: "AVAX", Icon: "https://example.com/yellow.png"},
	{Name: "Polygon", ShortName: "MATIC", Icon: "https://example.com/purple.png"},
}

func NewClient(cfg Config) *Client {
	// dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable", cfg.Host, cfg.User, cfg.Password, cfg.Database, cfg.Port)
	// gdb, err := gorm.Open(postgres.Open(dsn))
	// if err != nil {
	// 	panic(err)
	// }

	// err = gdb.AutoMigrate(&model.Game{}, &model.Player{}, &model.Camp{})
	// if err != nil {
	// 	panic(err)
	// }

	// err = gdb.Clauses(clause.OnConflict{
	// 	Columns:   []clause.Column{{Name: "name"}},
	// 	UpdateAll: true,
	// }).Create(&Camps).Error
	// if err != nil {
	// 	panic(err)
	// }

	// return &Client{DB: gdb, Game: game{DB: gdb}, Camp: camp{DB: gdb}, Player: player{DB: gdb}, Message: message{DB: gdb}}
	return &Client{}
}
