package game

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/COAOX/zecrey_warrior/config"
	"github.com/COAOX/zecrey_warrior/db"
	"github.com/COAOX/zecrey_warrior/model"
	"github.com/kvartborg/vector"
	"github.com/solarlune/resolv"
	"go.uber.org/zap"
)

type GameStatus int

const (
	CellTag           = "CELL"
	EdgeTag           = "EDGE"
	HorizontalEdgeTag = "HORIZONTAL"
	VerticalEdgeTag   = "VERTICAL"

	minCellSize = 5
	edgeWidth   = minCellSize + lineWidth

	playerInitialVelocity = 1

	GameNotStarted GameStatus = iota
	GameRunning
	GameStopped
)

type Game struct {
	db                *db.Client
	cfg               *config.Config
	onGameStart       func(context.Context)
	onGameStop        func(context.Context)
	onCampVotesChange func(camp Camp, votes int32)

	res *res

	space       *resolv.Space
	frameNumber uint32
	campVotes   sync.Map

	dbGame     *model.Game
	ctx        context.Context
	Map        Map `json:"map"`
	GameStatus GameStatus

	Players sync.Map `json:"players"`
	Items   sync.Map `json:"items"`

	nextRoundChan  chan struct{}
	stopSignalChan chan chan struct{}
}

func NewGame(ctx context.Context, cfg *config.Config, db *db.Client, onGameStart func(context.Context), onGameStop func(context.Context), onCampVotesChange func(camp Camp, votes int32)) *Game {
	v := &Game{
		ctx:               ctx,
		db:                db,
		cfg:               cfg,
		campVotes:         sync.Map{},
		Players:           sync.Map{},
		Items:             sync.Map{},
		onGameStart:       onGameStart,
		onGameStop:        onGameStop,
		onCampVotesChange: onCampVotesChange,
		GameStatus:        GameNotStarted,
		stopSignalChan:    make(chan chan struct{}, 1),
		nextRoundChan:     make(chan struct{}, 1),
	}

	zap.L().Debug("game init")

	v.initMap()
	v.initGameInfo()
	v.resetRes()

	// v.AddPlayer(11111, BTC)
	// v.AddPlayer(22222, ETH)
	// v.AddPlayer(33333, BNB)
	// v.AddPlayer(44444, AVAX)
	// v.AddPlayer(55555, MATIC)

	return v
}

func (g *Game) initMap() {
	g.Map = NewMap()

	g.space = resolv.NewSpace(int(g.Map.W())+2*edgeWidth, int(g.Map.H())+2*edgeWidth, edgeWidth, edgeWidth)
	g.space.Add(resolv.NewObject(0, 0, g.Map.W()+edgeWidth, edgeWidth, EdgeTag, HorizontalEdgeTag))
	g.space.Add(resolv.NewObject(0, edgeWidth, edgeWidth, g.Map.W()+edgeWidth, EdgeTag, VerticalEdgeTag))
	g.space.Add(resolv.NewObject(g.Map.W()+edgeWidth, 0, edgeWidth, g.Map.H()+edgeWidth, EdgeTag, VerticalEdgeTag))
	g.space.Add(resolv.NewObject(edgeWidth, g.Map.H()+edgeWidth, g.Map.W()+edgeWidth, edgeWidth, EdgeTag, HorizontalEdgeTag))

	for y := 0; y < mapRow; y++ {
		for x := 0; x < mapColumn; x++ {
			camp := initCamp(x, y)
			ox, oy := cellIndexToSpaceXY(x, y)
			g.space.Add(resolv.NewObject(ox, oy, float64(cellWidth), float64(cellHeight), CampTagMap[camp], CellTag, CellIndexToTag(x, y)))
			g.Map.Cells = append(g.Map.Cells, camp)
		}
	}
}

func (g *Game) initGameInfo() {
	g.dbGame = &model.Game{StartTime: time.Now(), EndTime: time.Now().Add(time.Duration(g.cfg.GameDuration) * time.Second)}
	if err := g.db.Game.Create(g.dbGame); err != nil {
		zap.L().Error("failed to create game", zap.Error(err))
	}
	zap.L().Debug("game info init", zap.Uint("game_id", g.dbGame.ID))
}

func (g *Game) resetRes() {
	g.res = nil
}

func (g *Game) GetGameID() uint {
	return g.dbGame.ID
}

func (g *Game) start() <-chan []byte {
	g.GameStatus = GameRunning
	stateChan := make(chan []byte)
	go func() {
		gameTime := time.NewTimer(time.Duration(g.cfg.GameDuration) * time.Second)
		for {
			s, _ := g.Serialize()
			g.Update()
			select {
			case <-g.ctx.Done():
				return
			case <-gameTime.C:
				g.nextRound()
				gameTime.Reset(time.Duration(g.cfg.GameDuration) * time.Second)
			default:
				stateChan <- s
			}
		}
	}()
	return stateChan
}

func (g *Game) nextRound() {
	g.Save()
	g.GameStatus = GameStopped
	g.stopSignalChan <- g.nextRoundChan
	g.onGameStop(g.ctx)
	// wait game to start
	<-time.After(time.Duration(g.cfg.GameRoundInterval) * time.Second)
	g.Reset()

	// g.AddPlayer(11111, BTC)
	// g.AddPlayer(22222, ETH)
	// g.AddPlayer(33333, BNB)
	// g.AddPlayer(44444, AVAX)
	// g.AddPlayer(55555, MATIC)

	g.onGameStart(g.ctx)
	g.nextRoundChan <- struct{}{}
}

// frame number: 4 bytes
// map size: 4 bytes
// map: map size bytes
// player number: 4 bytes
// players: 26 * len(players) bytes
// item number: 4 bytes
// items: 21 * items number bytes
func (g *Game) Serialize() ([]byte, error) {
	atomic.AddUint32(&g.frameNumber, 1)
	bytesBuf := bytes.NewBuffer([]byte{})
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, g.frameNumber)
	_, err := bytesBuf.Write(b)
	if err != nil {
		return bytesBuf.Bytes(), err
	}

	binary.BigEndian.PutUint32(b, g.Map.Size())
	bytesBuf.Write(b)
	bytesBuf.Write(g.Map.Serialize())

	playerNumber := uint32(0)
	playerBytes := []byte{}
	g.Players.Range(func(key, value interface{}) bool { // O(N) call, but since players are not that many, it's fine
		if v, ok := value.(*Player); ok && v != nil {
			playerNumber++
			playerBytes = append(playerBytes, v.Serialize()...)
		}
		return true
	})
	binary.BigEndian.PutUint32(b, playerNumber)
	bytesBuf.Write(b)
	bytesBuf.Write(playerBytes)

	itemNumber := uint32(0)
	itemBytes := []byte{}
	g.Items.Range(func(key, value interface{}) bool { // O(N) call, but since items are not that many, it's fine
		if v, ok := value.(*ItemObject); ok && v != nil {
			itemNumber++
			itemBytes = append(itemBytes, v.Serialize()...)
		}
		return true
	})
	binary.BigEndian.PutUint32(b, itemNumber)
	bytesBuf.Write(b)
	bytesBuf.Write(itemBytes)
	return bytesBuf.Bytes(), nil
}

func (g *Game) Save() {
	winner, _ := g.GetWinner()
	campID := uint8(winner)
	g.dbGame.WinnerID = campID
	g.dbGame.EndTime = time.Now()
	if err := g.db.Game.Update(g.dbGame); err != nil {
		zap.L().Error("failed to update game", zap.Error(err))
	}
	if err := g.db.Camp.IncreaseScore(campID); err != nil {
		zap.L().Error("failed to increase camp score", zap.Error(err))
	}
	if err := g.db.Player.IncreaseScore(g.dbGame.ID, campID); err != nil {
		zap.L().Error("failed to increase player score", zap.Error(err))
	}
}

func (g *Game) GetWinner() (Camp, int) {
	if g.res != nil {
		return g.res.winner, g.res.score
	}
	score := make(map[Camp]int)
	for _, v := range g.Map.Cells {
		score[v]++
	}
	maxScore := 0
	winner := BTC
	for k, v := range score {
		if v > maxScore && k != Empty {
			maxScore = v
			winner = k
		}
	}
	g.res = &res{winner: winner, score: maxScore}
	return winner, maxScore
}

func (g *Game) Reset() {
	g.Players = sync.Map{}
	g.campVotes = sync.Map{}
	g.Items = sync.Map{}
	g.frameNumber = 0
	g.initMap()
	g.initGameInfo()
	g.resetRes()
	g.GameStatus = GameRunning
}

func (g *Game) Update() {
	if g.GameStatus != GameRunning {
		return
	}
	g.Players.Range(func(key, value interface{}) bool {
		if player, ok := value.(*Player); ok && player != nil && player.playerObj != nil {
			remainX, remainY := player.Vx, player.Vy

			change := false
			for remainX != 0 || remainY != 0 {
				dx, dy := remainX, remainY
				if collision := player.playerObj.Check(dx, dy, getCollisionTags(player.Camp)...); collision != nil {
					collisionObj := collision.Objects[0]
					dx, dy = resolvDxDy(dx, dy, collision.ContactWithObject(collisionObj))
					if collisionObj.HasTags(CellTag) {
						remainX, remainY = player.rebound(dx, dy, remainX, remainY, collisionObj)
						if !change {
							change = true
							x, y := GetCellIndex(collisionObj.Tags())
							g.Map.Cells[y*mapColumn+x] = player.Camp
							collisionObj.RemoveTags(removeCampTags(collisionObj.Tags())...)
							collisionObj.AddTags(CampTagMap[player.Camp])
						}
					} else if collisionObj.HasTags(EdgeTag) {
						if collisionObj.HasTags(HorizontalEdgeTag) {
							player.Vy = -player.Vy
							remainX -= dx
							remainY = dy - remainY
						} else {
							player.Vx = -player.Vx
							remainX = dx - remainX
							remainY -= dy
						}
					} else if collisionObj.HasTags(ItemTag) {
						if collisionObj.HasTags(AcceleratorTag) {
							player.Vx *= 1.5
							player.Vy *= 1.5
							remainX *= 1.5
							remainY *= 1.5
							id := itemTagsToId(collisionObj.Tags())
							g.Items.Delete(id)
							g.space.Remove(collisionObj)
						}
					}
				} else {
					remainX -= dx
					remainY -= dy
				}
				// fmt.Println("#inner camp:", CampTagMap[player.Camp], "x:", player.playerObj.X, "y:", player.playerObj.Y, "dx:", dx, "dy:", dy, "vx:", player.Vx, "vy:", player.Vy, "rx:", remainX, "ry:", remainY)
				player.playerObj.X += dx
				player.playerObj.Y += dy
				player.playerObj.Update()
			}
			g.Players.Store(key, player)
		}
		return true
	})
	g.TryAddItem()
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

func (g *Game) incrCampVotes(camp Camp) {
	votes := int32(0)
	v, _ := g.campVotes.LoadOrStore(camp, &votes)
	i, ok := v.(*int32)
	if !ok {
		g.campVotes.Store(camp, &votes)
	}
	n := atomic.AddInt32(i, 1)
	g.onCampVotesChange(camp, n)
}

func GetCellIndex(tags []string) (int, int) {
	for _, tag := range tags {
		s := strings.Split(tag, ",")
		if len(s) == 2 {
			x, _ := strconv.Atoi(s[0])
			y, _ := strconv.Atoi(s[1])
			return x, y
		}
	}
	return 0, 0
}

func CellIndexToTag(x, y int) string {
	return fmt.Sprintf("%d,%d", x, y)
}

func CellTagToIndex(tag string) (int, int) {
	s := strings.Split(tag, ",")
	y, _ := strconv.Atoi(s[0])
	x, _ := strconv.Atoi(s[1])
	return x, y
}

func resolvDxDy(dx, dy float64, cvector vector.Vector) (x float64, y float64) {
	x, y = dx, dy
	cx, cy := cvector.X(), cvector.Y()
	xDistance, yDistance := float64(1), float64(1)
	if (cx < 0 && dx < cx) || (cx > 0 && dx > cx) {
		xDistance = cx / dx
	}
	if cx == 0 {
		if x == 0 {
			xDistance = 1
		} else {
			xDistance = 0
		}
	}

	if (cy < 0 && dy < cy) || (cy > 0 && dy > cy) {
		yDistance = cy / dy
	}
	if cy == 0 {
		if y == 0 {
			yDistance = 1
		} else {
			yDistance = 0
		}
	}

	if xDistance < yDistance {
		y *= xDistance
		x *= xDistance
	} else {
		x *= yDistance
		y *= yDistance
	}
	return
}

func map2SpaceXY(x, y float64) (float64, float64) {
	return x + edgeWidth, y + edgeWidth
}

func space2MapXY(x, y float64) (float64, float64) {
	return x - edgeWidth, y - edgeWidth
}

func cellIndexToSpaceXY(x, y int) (float64, float64) {
	return float64(x*(cellWidth+lineWidth) + edgeWidth), float64(y*(cellHeight+lineWidth) + edgeWidth)
}

type res struct {
	winner Camp
	score  int
}
