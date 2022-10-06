package game

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/COAOX/zecrey_warrior/config"
	"github.com/COAOX/zecrey_warrior/db"
	"github.com/topfreegames/pitaya/v2"
	"github.com/topfreegames/pitaya/v2/component"
)

const (
	gameRoomName = "game"
)

type Room struct {
	component.Base
	app pitaya.Pitaya
	cfg *config.Config
	db  *db.Client

	tickerCancel context.CancelFunc
	game         *Game
}

type GameUpdate struct {
	Data []byte `json:"data"`
}

func RegistRoom(app pitaya.Pitaya, db *db.Client, cfg *config.Config) *Game {
	err := app.GroupCreate(context.Background(), gameRoomName)
	if err != nil {
		panic(err)
	}
	r := &Room{
		app: app,
		db:  db,
		cfg: cfg,
	}
	r.game = NewGame(db, func(winner Camp) {
		r.onGameStop(context.Background(), winner)
	})
	app.Register(r,
		component.WithName(gameRoomName),
		component.WithNameFunc(strings.ToLower),
	)
	return r.game
}

func (r *Room) AfterInit() {
	ctx, cancel := context.WithCancel(context.Background())
	r.tickerCancel = cancel
	stateChan := r.game.Start(ctx)
	go func() {
		// ticker := time.Tick(time.Duration(33) * time.Millisecond)
		ticker := time.NewTicker(time.Duration(1000/r.cfg.FPS) * time.Millisecond).C
		for {
			select {
			case s := <-stateChan:
				<-ticker
				fmt.Println(s)
				r.app.GroupBroadcast(context.Background(), r.cfg.FrontendType, gameRoomName, "onUpdate", GameUpdate{Data: s})
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (r *Room) Shutdown() {
	r.tickerCancel()
}

// JoinResponse represents the result of joining room
type JoinResponse struct {
	Code   int    `json:"code"`
	Result string `json:"result"`
}

// UserMessage represents a message that user sent
type UserMessage struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

// NewUser message will be received when new user join room
type NewUser struct {
	Content string `json:"content"`
}

// Join room
func (r *Room) Join(ctx context.Context, msg []byte) (*JoinResponse, error) {
	s := r.app.GetSessionFromCtx(ctx)
	fakeUID := s.ID()                              // just use s.ID as uid !!!
	err := s.Bind(ctx, strconv.Itoa(int(fakeUID))) // binding session uid

	if err != nil {
		return nil, pitaya.Error(err, "RH-000", map[string]string{"failed": "bind"})
	}

	// uids, err := r.app.GroupMembers(ctx, gameRoomName)
	// if err != nil {
	// 	return nil, err
	// }
	// s.Push("onMembers", &AllMembers{Members: uids})

	// new user join group
	r.app.GroupAddMember(ctx, gameRoomName, s.UID()) // add session to group

	// notify others
	r.onJoin(ctx)

	// on session close, remove it from group
	s.OnClose(func() {
		r.app.GroupRemoveMember(ctx, gameRoomName, s.UID())
	})

	return &JoinResponse{Result: "success"}, nil
}

// Message sync last message to all members
func (r *Room) Message(ctx context.Context, msg *UserMessage) {
	// fmt.Println("Message: ", msg)
	err := r.app.GroupBroadcast(ctx, r.cfg.FrontendType, gameRoomName, "onMessage", msg)
	if err != nil {
		// fmt.Println("error broadcasting message", err)
	}
}

func (r *Room) onJoin(ctx context.Context) {
	gi := GameInfo{
		Row:        r.game.Map.Row,
		Column:     r.game.Map.Column,
		CellWidth:  r.game.Map.CellWidth,
		CellHeight: r.game.Map.CellHeight,
	}
	r.game.Players.Range(func(key, value interface{}) bool {
		if p, ok := value.(*Player); ok {
			gi.Players = append(gi.Players, PlayerInfo{
				ID:        p.ID,
				Thumbnail: p.Thumbnail,
			})
		}
		return true
	})

	r.app.GroupBroadcast(ctx, r.cfg.FrontendType, gameRoomName, "onJoin", gi)
}

func (r *Room) onGameStop(ctx context.Context, winer Camp) {
	r.app.GroupBroadcast(ctx, r.cfg.FrontendType, gameRoomName, "onGameStop", GameStop{
		Winner:        winer,
		NextCountDown: int64(r.cfg.GameRoundInterval),
	})
	r.app.GroupRemoveAll(ctx, gameRoomName)
}

// TODO
type GameInfo struct {
	Row    uint32 `json:"row"`
	Column uint32 `json:"column"`

	CellWidth  uint32 `json:"cell_width"`
	CellHeight uint32 `json:"cell_height"`

	Players []PlayerInfo `json:"players"`
}

type PlayerInfo struct {
	ID        uint64 `json:"id"`
	Thumbnail string `json:"thumbnail"`
}

type GameStop struct {
	Winner        Camp  `json:"winner"`
	NextCountDown int64 `json:"next_count_down"`
}
