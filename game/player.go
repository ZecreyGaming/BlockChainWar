package game

import (
	"bytes"
	"encoding/binary"
	"math"

	"github.com/solarlune/resolv"
)

const (
	PlayerTag           = "Player"
	defaultPlayerPixelR = 15
)

type Player struct {
	ID        uint64 `json:"player_id"`
	Camp      Camp   `json:"camp"`
	Thumbnail string `json:"thumbnail"`

	R int `json:"r"`

	Vx float64 `json:"vx"`
	Vy float64 `json:"vy"`

	playerObj *resolv.Object
}

// ID 8 byte
// R 2 byte
// X 8 byte
// Y 8 byte
func (p *Player) Serialize() []byte {
	bytesBuffer := bytes.NewBuffer(make([]byte, 0))
	binary.Write(bytesBuffer, binary.BigEndian, p.ID)
	binary.Write(bytesBuffer, binary.BigEndian, uint16(p.R))
	x, y := float64(0), float64(0)
	if p.playerObj != nil {
		x, y = p.playerObj.X-edgeWidth, p.playerObj.Y-edgeWidth
	}
	binary.Write(bytesBuffer, binary.BigEndian, x)
	binary.Write(bytesBuffer, binary.BigEndian, y)
	return bytesBuffer.Bytes()
}

func (p *Player) Size() uint32 {
	return 18
}

func (player *Player) rebound(dx, dy, rx, ry float64, cell *resolv.Object) (float64, float64) {
	// Edge Collision
	nx, ny := player.playerObj.X+dx, player.playerObj.Y+dy
	rx -= dx
	ry -= dy
	if nx >= cell.X && nx <= cell.X+cell.W {
		player.Vy = -player.Vy
		return rx, -ry
	}
	if ny >= cell.Y && ny <= cell.Y+cell.H {
		player.Vx = -player.Vx
		return -rx, ry
	}

	// Corner Collision
	remianV := math.Sqrt(rx*rx + ry*ry)
	if remianV == 0 {
		return 0, 0
	}

	v := math.Sqrt(player.Vx*player.Vx + player.Vy*player.Vy)
	player.Vx, player.Vy = v/float64(player.R)*(nx-cell.X), v/float64(player.R)*(ny-cell.Y)

	px, py := (nx - cell.X), (ny - cell.Y)
	pl := math.Sqrt(px*px + py*py)
	return remianV / pl * px, remianV / pl * py
}
