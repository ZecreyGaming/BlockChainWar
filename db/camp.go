package db

import (
	"github.com/COAOX/zecrey_warrior/model"
	"gorm.io/gorm"
)

type camp db

func (c *camp) Create(camp *model.Camp) error {
	return c.db.Create(camp).Error
}

func (c *camp) IncreaseScore(campID uint8) error {
	return c.db.Model(&model.Camp{}).Where("id = ?", campID).Update("score", gorm.Expr("score + ?", 1)).Error
}
