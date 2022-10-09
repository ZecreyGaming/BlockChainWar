package db

import (
	"github.com/COAOX/zecrey_warrior/model"
	"gorm.io/gorm"
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

func (p *player) List(playeIDs ...uint64) ([]model.Player, error) {
	var players []model.Player
	err := p.db.Where("player_id in ?", playeIDs).Find(&players).Error
	return players, err
}

func (p *player) ListRank(limit int) ([]model.Player, error) {
	var players []model.Player
	err := p.db.Order("score desc").Limit(limit).Find(&players).Error
	return players, err
}

func (p *player) IncreaseScore(gameID uint, campID uint8) error {
	player_votes := []model.PlayerVote{}
	if err := p.db.Model(&model.PlayerVote{}).Where("game_id = ? AND camp = ?", gameID, campID).Find(&player_votes).Error; err != nil {
		return err
	}
	playerIds := []uint64{}
	for _, player_vote := range player_votes {
		playerIds = append(playerIds, player_vote.PlayerID)
	}
	return p.db.Model(&model.Player{}).Where("player_id in ?", playerIds).Update("score", gorm.Expr("score + ?", 1)).Error
}

func (p *player) AddVote(playerVotes *model.PlayerVote) error {
	return p.db.Create(playerVotes).Error
}

func (p *player) GetWinnerVotes(gameId uint, winner uint8) int64 {
	var count int64
	p.db.Model(&model.PlayerVote{}).Where("game_id = ? AND camp = ?", gameId, winner).Count(&count)
	return count
}
