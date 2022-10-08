package game

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"math/rand"

	"github.com/solarlune/resolv"
)

const (
	PlayerTag           = "Player"
	defaultPlayerPixelR = 5
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
		x, y = space2MapXY(p.GetCenter())
	}
	binary.Write(bytesBuffer, binary.BigEndian, x)
	binary.Write(bytesBuffer, binary.BigEndian, y)
	return bytesBuffer.Bytes()
}

func (p *Player) Size() uint32 {
	return 18
}

func (p *Player) GetCenter() (float64, float64) {
	return p.playerObj.X + float64(p.R), p.playerObj.Y + float64(p.R)
}

func (player *Player) rebound(dx, dy, rx, ry float64, cell *resolv.Object) (float64, float64) {
	// Edge Collision
	nx, ny := player.playerObj.X+dx, player.playerObj.Y+dy
	rx -= dx
	ry -= dy
	// if nx >= cell.X && nx <= cell.X+cell.W {
	// 	player.Vy = -player.Vy
	// 	return rx, -ry
	// }
	// if ny >= cell.Y && ny <= cell.Y+cell.H {
	// 	player.Vx = -player.Vx
	// 	return -rx, ry
	// }

	// // Corner Collision
	// remianV := math.Sqrt(rx*rx + ry*ry)
	// if remianV == 0 {
	// 	return 0, 0
	// }

	// v := math.Sqrt(player.Vx*player.Vx + player.Vy*player.Vy)
	// player.Vx, player.Vy = v/float64(player.R)*(cell.X-nx), v/float64(player.R)*(cell.Y-ny)

	// px, py := (cell.X - nx), (cell.Y - ny)
	// pl := math.Sqrt(px*px + py*py)
	// player.Vx *= -1
	// player.Vy *= -1
	// return -rx, -ry

	if ny <= cell.Y-float64(2*player.R) || ny >= cell.Y+cell.H {
		player.Vy *= -1
		ry *= -1
	}

	if nx <= cell.X-float64(2*player.R) || nx >= cell.X+cell.W {
		player.Vx *= -1
		rx *= -1
	}
	return rx, ry
}

func (g *Game) AddPlayer(playerID uint64, camp Camp) *Player {
	// if g.GameStatus != GameRunning {
	// 	return nil
	// }
	if camp == Empty {
		return nil
	}
	g.incrCampVotes(camp)
	x, y := cellIndexToSpaceXY(camp.CenterCellIndex(mapRow, mapColumn))

	ang := rand.Float64() * 2 * math.Pi
	player := &Player{
		ID:   playerID,
		Camp: camp,
		R:    defaultPlayerPixelR,
		Vx:   math.Cos(ang) * playerInitialVelocity,
		Vy:   math.Sin(ang) * playerInitialVelocity,
	}
	player.playerObj = resolv.NewObject(x, y, float64(2*player.R), float64(2*player.R), PlayerTag)
	g.space.Add(player.playerObj)
	g.Players.Store(playerID, player)

	fmt.Println("new player, camp:", camp, "x:", player.playerObj.X, "y:", player.playerObj.Y, "vx:", player.Vx, "vy:", player.Vy)
	return player
}
