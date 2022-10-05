package db

import "github.com/COAOX/zecrey_warrior/model"

type player db

func (p *player) Create(player *model.Player) error {
	return p.DB.Create(player).Error
}
