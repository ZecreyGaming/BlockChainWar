package game

import "github.com/ZecreyGaming/BlockChainWar/model"

type GameInfo struct {
	*model.Game
	GameRound      uint            `json:"game_round"`
	HistoryMessage []model.Message `json:"history_message"`
	CampVotes      map[Camp]int32  `json:"camp_votes"`
	CampRank       []model.Camp    `json:"camp_rank"`
	PlayerRank     []model.Player  `json:"player_rank"`
	WinnerId       uint8           `json:"winner_id"`
	GameStatus     GameStatus      `json:"game_status"` //0 1 2 : 没开始，进行中，已结束
}

func (g *Game) GetGameInfo() (GameInfo, error) {
	var err error
	var GameRound uint
	if g.dbGame != nil {
		GameRound = g.dbGame.ID
	}

	v := GameInfo{
		Game:      g.dbGame,
		GameRound: GameRound,
		CampVotes: map[Camp]int32{},
	}
	offset, limit := 0, 100
	v.HistoryMessage, err = g.db.Message.ListLatest(offset, limit)
	if err != nil {
		return v, err
	}
	g.campVotes.Range(func(key, value interface{}) bool {
		if c, ok := key.(Camp); ok && value.(*int32) != nil {
			v.CampVotes[c] = *(value.(*int32))
		}
		return true
	})

	rankLimit := 3
	v.CampRank, err = g.db.Camp.ListRank(rankLimit)
	if err != nil {
		return v, err
	}

	v.PlayerRank, err = g.db.Player.ListRank(rankLimit)
	if err != nil {
		return v, err
	}
	v.GameStatus = g.GameStatus
	winnerId, _ := g.GetLastWinner()
	v.WinnerId = winnerId
	return v, nil
}

type GameStop struct {
	Winner        Camp           `json:"winner"`
	WinnerVotes   int64          `json:"winner_votes"`
	NextCountDown int64          `json:"next_count_down"`
	CampRank      []model.Camp   `json:"camp_rank"`
	PlayerRank    []model.Player `json:"player_rank"`
}

func (g *Game) GetGameStop() GameStop {
	winner, _ := g.GetLastWinner()
	v := GameStop{
		Winner:        Camp(winner),
		WinnerVotes:   g.db.Player.GetWinnerVotes(g.dbGame.ID, uint8(winner)),
		NextCountDown: int64(g.cfg.GameRoundInterval),
	}
	rankLimit := 3
	v.CampRank, _ = g.db.Camp.ListRank(rankLimit)
	v.PlayerRank, _ = g.db.Player.ListRank(rankLimit)
	return v
}
