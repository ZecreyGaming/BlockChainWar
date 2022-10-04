package model

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/solarlune/resolv"
)

const (
	sizeOfCellStateBits = 4
	campMaskLeft        = byte(0xF0)
	campMaskRight       = byte(0x0F)
)

type Camp uint8 // should convert to int4 when transfered to client

const (
	Empty Camp = iota
	BTC
	ETH
	BNB
	AVAX
	MATIC

	BTCInitialLen   = 6
	ETHInitialLen   = 5
	BNBInitialLen   = 4
	AVAXInitialLen  = 3
	MATICInitialLen = 2

	EmptyTag = "Empty"
	BTCTag   = "BTC"
	ETHTag   = "ETH"
	BNBTag   = "BNB"
	AVAXTag  = "AVAX"
	MATICTag = "MATIC"

	EdgeTag           = "EDGE"
	HorizontalEdgeTag = "HORIZONTAL"
	VerticalEdgeTag   = "VERTICAL"

	PlayerTag           = "Player"
	defaultPlayerPixelR = 5

	minCellSize = defaultPlayerPixelR
	edgeWidth   = defaultPlayerPixelR
)

var (
	CampTagMap = map[Camp]string{
		Empty: EmptyTag,
		BTC:   BTCTag,
		ETH:   ETHTag,
		BNB:   BNBTag,
		AVAX:  AVAXTag,
		MATIC: MATICTag,
	}

	CampTagMapReverse = map[string]Camp{
		EmptyTag: Empty,
		BTCTag:   BTC,
		ETHTag:   ETH,
		BNBTag:   BNB,
		AVAXTag:  AVAX,
		MATICTag: MATIC,
	}
)

type Map struct {
	Row    uint32 `json:"row"`
	Column uint32 `json:"column"`

	CellWidth  uint32 `json:"cell_width"`
	CellHeight uint32 `json:"cell_height"`
	LineWidth  uint32 `json:"line_width"`

	Cells []Camp `json:"cells"`
}

func (m *Map) Serialize() []byte {
	l := 20 + len(m.Cells)*sizeOfCellStateBits/8
	res := make([]byte, l)
	binary.LittleEndian.PutUint32(res[0:4], m.Row)
	binary.LittleEndian.PutUint32(res[4:8], m.Column)
	binary.LittleEndian.PutUint32(res[8:12], m.CellWidth)
	binary.LittleEndian.PutUint32(res[12:16], m.CellHeight)
	binary.LittleEndian.PutUint32(res[16:20], m.LineWidth)
	offset := 20
	for i := 0; i < len(m.Cells); i += 2 {
		n := byte(m.Cells[i]<<4) & campMaskLeft
		if i+1 < len(m.Cells) {
			n = n | (byte(m.Cells[i+1]) & campMaskRight)
		}
		res[offset] = n
		offset++
	}
	return res
}

func (m *Map) Size() uint32 {
	return uint32(20 + len(m.Cells)*sizeOfCellStateBits/8)
}

type Player struct {
	PlayerID uint64  `json:"player_id"`
	Camp     Camp    `json:"camp"`
	X        float64 `json:"px"`
	Y        float64 `json:"py"`

	R uint8 `json:"r"`

	Vx float64 `json:"vx"`
	Vy float64 `json:"vy"`

	playerObj *resolv.Object
}

func (p *Player) Serialize() []byte {
	b := make([]byte, 18)
	binary.LittleEndian.PutUint64(b[0:8], p.PlayerID)
	b[8] = byte(p.Camp)
	b[9] = byte(p.R)
	binary.LittleEndian.PutUint32(b[10:14], uint32(p.X))
	binary.LittleEndian.PutUint32(b[14:18], uint32(p.Y))
	return b
}

func (p *Player) Size() uint32 {
	return 18
}

type Game struct {
	Map     Map      `json:"map"`
	Players sync.Map `json:"players"`

	space *resolv.Space

	frameNumber uint32
}

// frame number: 4 bytes
// map size: 4 bytes
// map: map size bytes
// players: 18 * len(players) bytes
func (g *Game) Serialize() ([]byte, error) {
	atomic.AddUint32(&g.frameNumber, 1)
	bytesBuf := bytes.NewBuffer([]byte{})
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, g.Size())
	_, err := bytesBuf.Write(b)
	if err != nil {
		return bytesBuf.Bytes(), err
	}
	binary.LittleEndian.PutUint32(b, g.frameNumber)
	_, err = bytesBuf.Write(b)
	if err != nil {
		return bytesBuf.Bytes(), err
	}

	binary.LittleEndian.PutUint32(b, g.Map.Size())
	bytesBuf.Write(b)
	bytesBuf.Write(g.Map.Serialize())

	g.Players.Range(func(key, value interface{}) bool {
		if v, ok := value.(*Player); ok && v != nil {
			bytesBuf.Write(v.Serialize())
		}
		return true
	})

	// by, _ := json.Marshal(g)
	// fmt.Println("game", string(by))
	// fmt.Println("cells", g.Map.Cells)

	return bytesBuf.Bytes(), nil
}

func (g *Game) Update() {
	g.Players.Range(func(key, value interface{}) bool {
		if player, ok := value.(*Player); ok && player != nil {
			if player.playerObj == nil {
				player.playerObj = resolv.NewObject(player.X-float64(player.R), player.Y-float64(player.R), float64(2*player.R), float64(2*player.R), PlayerTag)
				player.playerObj.SetShape(resolv.NewCircle(float64(player.R), float64(player.R), float64(player.R)))
				g.space.Add(player.playerObj)
			}
			remainX, remainY := float64(player.Vx), float64(player.Vy)
			for remainX != 0 || remainY != 0 {
				dx, dy := remainX, remainY
				if collision := player.playerObj.Check(dx, dy, getCollisionTags(player.Camp)...); collision != nil {
					collisionObj := collision.Objects[0]
					dx = collision.ContactWithObject(collisionObj).X()
					dy = collision.ContactWithObject(collisionObj).Y()
					fmt.Println("########")
					fmt.Println("collision, has tag: ", collisionObj.HasTags(EdgeTag))
					if !collisionObj.HasTags(EdgeTag) {
						remainX, remainY = player.rebound(dx, dy, remainX, remainY, collisionObj)
						x, y := GetCellIndex(collisionObj.Tags())
						g.Map.Cells[y*int(g.Map.Column)+x] = player.Camp
						collisionObj.RemoveTags(removeCampTags(collisionObj.Tags())...)
						collisionObj.AddTags(CampTagMap[player.Camp])
					} else if collisionObj.HasTags(HorizontalEdgeTag) {
						player.Vx = -player.Vx
						remainX = dx - remainX
						remainY -= dy
					} else {
						player.Vy = -player.Vy
						remainX -= dx
						remainY = dy - remainY
					}
				} else {
					remainX -= dx
					remainY -= dy
				}
				fmt.Println("x:", player.playerObj.X, "y:", player.playerObj.Y, "dx:", dx, "dy:", dy, "rx", remainX, "ry", remainY)
				player.playerObj.X += dx
				player.playerObj.Y += dy
				player.playerObj.Update()
				player.X += dx
				player.Y += dy
			}

			g.Players.Store(key, player)
		}
		return true
	})
}

func (g *Game) Size() uint32 {
	pLen := uint32(0)
	g.Players.Range(func(key, value interface{}) bool { // O(N) call, but since players are not that many, it's fine
		if v, ok := value.(*Player); ok && v != nil {
			pLen += v.Size()
		}
		return true
	})
	return 4 + 4 + g.Map.Size() + pLen
}

func NewGame() *Game {
	v := &Game{
		Map: Map{
			Row:        30,
			Column:     40,
			Cells:      []Camp{},
			CellWidth:  20,
			CellHeight: 20,
			LineWidth:  1,
		},
		Players: sync.Map{},
	}
	//TODO
	v.Players.Store("test_player", &Player{
		PlayerID: 1,
		Camp:     ETH,
		X:        edgeWidth + defaultPlayerPixelR,
		Y:        edgeWidth + defaultPlayerPixelR + 10,
		R:        defaultPlayerPixelR,
		Vx:       0,
		Vy:       -5,
	})

	v.space = resolv.NewSpace(int(v.Map.Column*v.Map.CellWidth)+2*edgeWidth, int(v.Map.Row*v.Map.CellHeight)+2*edgeWidth, minCellSize, minCellSize)
	v.space.Add(resolv.NewObject(0, 0, float64(v.Map.Column*v.Map.CellWidth+edgeWidth), edgeWidth, EdgeTag, HorizontalEdgeTag))
	v.space.Add(resolv.NewObject(0, 0, edgeWidth, float64(v.Map.Column*v.Map.CellWidth+edgeWidth), EdgeTag, VerticalEdgeTag))
	v.space.Add(resolv.NewObject(float64(v.Map.Column*v.Map.CellWidth+edgeWidth), 0, edgeWidth, float64(v.Map.Row*v.Map.CellHeight+edgeWidth), EdgeTag, VerticalEdgeTag))
	v.space.Add(resolv.NewObject(0, float64(v.Map.Row*v.Map.CellHeight+edgeWidth), float64(v.Map.Column*v.Map.CellWidth+edgeWidth), edgeWidth, EdgeTag, HorizontalEdgeTag))

	for i := 0; i < int(v.Map.Row); i++ {
		for j := 0; j < int(v.Map.Column); j++ {
			camp := initCamp(i, j, int(v.Map.Row), int(v.Map.Column))
			v.space.Add(resolv.NewObject(float64(j*int(v.Map.CellWidth)), float64(i*int(v.Map.CellHeight)), float64(v.Map.CellWidth), float64(v.Map.CellHeight), CampTagMap[camp], CellIndexToTag(i, j)))
			v.Map.Cells = append(v.Map.Cells, camp)
		}
	}

	return v
}

func initCamp(i, j, r, c int) Camp {
	if i >= 0 && i < ETHInitialLen && j >= 0 && j < ETHInitialLen {
		return ETH
	}

	if i >= 0 && i < BNBInitialLen && j < c && j >= c-BNBInitialLen {
		return BNB
	}

	if i >= (r-BTCInitialLen)/2 && i < (r+BTCInitialLen)/2 && j >= (c-BTCInitialLen)/2 && j < (c+BTCInitialLen)/2 {
		return BTC
	}

	if i >= r-AVAXInitialLen && i < r && j >= 0 && j < AVAXInitialLen {
		return AVAX
	}

	if i >= r-MATICInitialLen && i < r && j >= c-MATICInitialLen && j < c {
		return MATIC
	}
	return Empty
}

func GetCellIndex(tags []string) (int, int) {
	for _, tag := range tags {
		s := strings.Split(tag, ",")
		if len(s) == 2 {
			y, _ := strconv.Atoi(s[0])
			x, _ := strconv.Atoi(s[1])
			return x, y
		}
	}
	return 0, 0
}

func CellIndexToTag(x, y int) string {
	return fmt.Sprintf("%d,%d", y, x)
}

func CellTagToIndex(tag string) (int, int) {
	s := strings.Split(tag, ",")
	y, _ := strconv.Atoi(s[0])
	x, _ := strconv.Atoi(s[1])
	return x, y
}

func getCollisionTags(camp Camp) (retval []string) {
	switch camp {
	case BTC:
		retval = []string{CampTagMap[ETH], CampTagMap[BNB], CampTagMap[AVAX], CampTagMap[MATIC], CampTagMap[Empty]}
	case ETH:
		retval = []string{CampTagMap[BNB], CampTagMap[BTC], CampTagMap[AVAX], CampTagMap[MATIC], CampTagMap[Empty]}
	case BNB:
		retval = []string{CampTagMap[ETH], CampTagMap[BTC], CampTagMap[AVAX], CampTagMap[MATIC], CampTagMap[Empty]}
	case AVAX:
		retval = []string{CampTagMap[ETH], CampTagMap[BNB], CampTagMap[BTC], CampTagMap[MATIC], CampTagMap[Empty]}
	case MATIC:
		retval = []string{CampTagMap[ETH], CampTagMap[BNB], CampTagMap[BTC], CampTagMap[AVAX], CampTagMap[Empty]}
	default:
		retval = []string{CampTagMap[BTC], CampTagMap[ETH], CampTagMap[BNB], CampTagMap[AVAX], CampTagMap[MATIC], CampTagMap[Empty]}
	}
	retval = append(retval, HorizontalEdgeTag, VerticalEdgeTag, EdgeTag)
	return
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
	l := math.Sqrt(rx*rx + ry*ry)
	if l == 0 {
		return 0, 0
	}

	v := math.Sqrt(player.Vx*player.Vx + player.Vy*player.Vy)
	player.Vx, player.Vy = v/float64(player.R)*(nx-cell.X), v/float64(player.R)*(ny-cell.Y)

	return l / float64(player.R) * (nx - cell.X), l / float64(player.R) * (ny - cell.Y)
}

func removeCampTags(tags []string) []string {
	ret := []string{}
	for _, tag := range tags {
		if _, ok := CampTagMapReverse[tag]; ok {
			ret = append(ret, tag)
		}
	}
	return ret
}
