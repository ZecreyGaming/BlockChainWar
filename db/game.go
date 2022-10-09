package db

import (
	"fmt"

	"github.com/COAOX/zecrey_warrior/model"
)

type game db

func (g *game) Create(game *model.Game) error {
	fmt.Println("on create game")
	return g.db.Create(game).Error
}

func (g *game) Update(game *model.Game) error {
	return g.db.Updates(game).Error
}
