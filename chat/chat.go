package chat

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/COAOX/zecrey_warrior/config"
	"github.com/COAOX/zecrey_warrior/db"
	"github.com/COAOX/zecrey_warrior/game"
	"github.com/COAOX/zecrey_warrior/model"
	"github.com/topfreegames/pitaya/constants"
	"github.com/topfreegames/pitaya/v2"
	"github.com/topfreegames/pitaya/v2/component"
	"go.uber.org/zap"
)

const (
	chatRoomName = "chat"
)

type Room struct {
	component.Base
	app pitaya.Pitaya
	cfg *config.Config
	db  *db.Client

	game *game.Game
}

func RegistRoom(app pitaya.Pitaya, db *db.Client, cfg *config.Config, game *game.Game) {
	err := app.GroupCreate(context.Background(), chatRoomName)
	if err != nil {
		panic(err)
	}

	app.Register(&Room{
		app:  app,
		db:   db,
		cfg:  cfg,
		game: game,
	},
		component.WithName(chatRoomName),
		component.WithNameFunc(strings.ToLower),
	)
}

// JoinResponse represents the result of joining room
type JoinResponse struct {
	Code        int       `json:"code"`
	Result      string    `json:"result"`
	GameRound   uint      `json:"game_round"`
	GameEndTime time.Time `json:"game_end_time"`
}

type MessageResponse struct {
	Code   int    `json:"code"`
	Result string `json:"result"`
}

// NewUser message will be received when new user join room
type NewUser struct {
	Content string `json:"content"`
}

// Join room
func (r *Room) Join(ctx context.Context, msg []byte) (*JoinResponse, error) {
	fmt.Println("on chat join")
	s := r.app.GetSessionFromCtx(ctx)
	fakeUID := s.ID()                              // just use s.ID as uid !!!
	err := s.Bind(ctx, strconv.Itoa(int(fakeUID))) // binding session uid

	if err != nil && err != constants.ErrSessionAlreadyBound {
		return nil, pitaya.Error(err, "RH-000", map[string]string{"failed": "bind"})
	}

	offset, limit := 0, 100
	// get last 30 messages
	messages, err := r.db.Message.ListLatest(offset, limit)
	if err != nil {
		return nil, pitaya.Error(err, "RH-500", map[string]string{"failed": "get messages"})
	}
	s.Push("onHistoryMessage", messages)

	// uids, err := r.app.GroupMembers(ctx, gameRoomName)
	// if err != nil {
	// 	return nil, err
	// }
	// s.Push("onMembers", &AllMembers{Members: uids})

	// new user join group
	r.app.GroupAddMember(ctx, chatRoomName, s.UID()) // add session to group

	// on session close, remove it from group
	s.OnClose(func() {
		r.app.GroupRemoveMember(ctx, chatRoomName, s.UID())
	})

	return &JoinResponse{Result: "success", GameRound: r.game.ID}, nil
}

// Message sync last message to all members
func (r *Room) Message(ctx context.Context, msg *model.Message) (*MessageResponse, error) {
	err := r.app.GroupBroadcast(ctx, r.cfg.FrontendType, chatRoomName, "onMessage", msg)
	if err != nil {
		zap.L().Error("broadcast message failed", zap.Error(err))
	}
	err = r.db.Message.Create(&model.Message{})
	if err != nil {
		zap.L().Error("save message failed", zap.Error(err))
	}

	if r.game != nil {
		r.game.AddPlayer(msg.PlayerID, game.DecideCamp(msg.Message))
	}
	return &MessageResponse{
		Result: "success",
	}, nil
}
