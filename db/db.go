package db

import (
	"fmt"

	"github.com/COAOX/zecrey_warrior/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	db *gorm.DB
}

func NewClient(cfg Config) *Client {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable", cfg.Host, cfg.User, cfg.Password, cfg.Database, cfg.Port)
	gdb, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		panic(err)
	}

	err = gdb.AutoMigrate(&model.Message{}, &model.Game{}, &model.Player{}, &model.Camp{}, &model.PlayerVote{})
	if err != nil {
		panic(err)
	}

	err = gdb.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}},
		UpdateAll: true,
	}).Create(&model.Camps).Error
	if err != nil {
		panic(err)
	}

	return &Client{DB: gdb, Game: game{db: gdb}, Camp: camp{db: gdb}, Player: player{db: gdb}, Message: message{db: gdb}}
	// return &Client{}
}
