package db

import (
	"github.com/COAOX/zecrey_warrior/model"
	"gorm.io/gorm/clause"
)

type player db

func (p *player) Create(player *model.Player) error {
	return p.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "player_id"}},
		UpdateAll: true,
	}).Create(player).Error
}

func (p *player) Get(playerID uint64) (model.Player, error) {
	var player model.Player
	err := p.db.First(&player, "player_id = ?", playerID).Error
	return player, err
}
