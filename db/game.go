package db

import (
	"github.com/COAOX/zecrey_warrior/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type game db

func (g *game) Create(game *model.Game) error {
	return g.db.Create(game).Error
}

func (g *game) Update(game *model.Game) error {
	return g.db.Updates(game).Error
}
func (g *game) GetLastWinner() (*model.Game, error) {
	var _game *model.Game
	db := g.db.Preload(clause.Associations).Order(clause.OrderByColumn{Column: clause.Column{Name: "created_at"}, Desc: true}).Limit(1).Find(&_game)
	if db.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return _game, db.Error
}
