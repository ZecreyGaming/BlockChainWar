package db

import "github.com/COAOX/zecrey_warrior/model"

type message db

func (m *message) Create(message *model.Message) error {
	return m.DB.Create(message).Error
}

func (m *message) ListMessages(gameID uint) []model.Message {
	var messages []model.Message
	m.DB.Where("game_id = ?", gameID).Find(&messages)
	return messages
}
