package game

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/COAOX/zecrey_warrior/config"
	"github.com/COAOX/zecrey_warrior/db"
	"github.com/COAOX/zecrey_warrior/state"
	"github.com/topfreegames/pitaya/v2"
	"github.com/topfreegames/pitaya/v2/component"
)

type Room struct {
	component.Base
	app pitaya.Pitaya
	cfg *config.Config
	db  *db.Client

	tickerCancel context.CancelFunc
	game         *state.Game
}

type GameUpdate struct {
	Data []byte `json:"data"`
}

func NewRoom(app pitaya.Pitaya, db *db.Client, cfg *config.Config) *Room {
	return &Room{
		app:  app,
		game: state.NewGame(db),
		db:   db,
		cfg:  cfg,
	}
}

func (r *Room) AfterInit() {
	ctx, cancel := context.WithCancel(context.Background())
	r.tickerCancel = cancel
	stateChan := make(chan []byte)
	go func() {
		for {
			s, _ := r.game.Serialize()
			r.game.Update()
			select {
			case <-ctx.Done():
				return
			case stateChan <- s:
			}
		}
	}()
	go func() {
		// ticker := time.Tick(time.Duration(33) * time.Millisecond)
		ticker := time.NewTicker(time.Duration(1000/r.cfg.FPS) * time.Millisecond).C
		for {
			select {
			case s := <-stateChan:
				<-ticker
				fmt.Println(s)
				r.app.GroupBroadcast(context.Background(), "zecrey_warrior", "room", "onUpdate", GameUpdate{Data: s})
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

	uids, err := r.app.GroupMembers(ctx, "room")
	if err != nil {
		return nil, err
	}
	s.Push("onMembers", &AllMembers{Members: uids})
	// notify others
	r.app.GroupBroadcast(ctx, "zecrey_warrior", "room", "onNewUser", &NewUser{Content: fmt.Sprintf("New user: %s", s.UID())})
	// new user join group
	r.app.GroupAddMember(ctx, "room", s.UID()) // add session to group

	// on session close, remove it from group
	s.OnClose(func() {
		r.app.GroupRemoveMember(ctx, "room", s.UID())
	})

	return &JoinResponse{Result: "success"}, nil
}

// Message sync last message to all members
func (r *Room) Message(ctx context.Context, msg *UserMessage) {
	// fmt.Println("Message: ", msg)
	err := r.app.GroupBroadcast(ctx, "zecrey_warrior", "room", "onMessage", msg)
	if err != nil {
		// fmt.Println("error broadcasting message", err)
	}
}
