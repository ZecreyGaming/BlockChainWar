package model

import (
	"encoding/binary"

	"github.com/solarlune/resolv"
)

const (
	PlayerTag           = "Player"
	defaultPlayerPixelR = minCellSize
)

type Player struct {
	ID   uint64 `json:"player_id"`
	Camp Camp   `json:"camp"`

	R int `json:"r"`

	Vx float64 `json:"vx"`
	Vy float64 `json:"vy"`

	playerObj *resolv.Object
}

func (p *Player) Serialize() []byte {
	b := make([]byte, 18)
	binary.LittleEndian.PutUint64(b[0:8], p.ID)
	b[8] = byte(p.Camp)
	b[9] = byte(p.R)
	if p.playerObj != nil {
		binary.LittleEndian.PutUint32(b[10:14], uint32(p.playerObj.X))
		binary.LittleEndian.PutUint32(b[14:18], uint32(p.playerObj.Y))
	}
	return b
}

func (p *Player) Size() uint32 {
	return 18
}
