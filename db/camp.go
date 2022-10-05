package db

import "github.com/COAOX/zecrey_warrior/model"

type camp db

func (c *camp) Create(camp *model.Camp) error {
	return c.DB.Create(camp).Error
}
