package db

import "github.com/COAOX/zecrey_warrior/model"

type game db

func (g *game) Create(game *model.Game) error {
	return g.DB.Create(game).Error
}
