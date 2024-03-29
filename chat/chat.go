package chat

import (
	"context"
	"fmt"
	sdk "github.com/ZecreyGaming/BlockChainWar/game/cronjob/zecreyface"
	"strconv"
	"strings"

	"github.com/ZecreyGaming/BlockChainWar/config"
	"github.com/ZecreyGaming/BlockChainWar/db"
	"github.com/ZecreyGaming/BlockChainWar/game"
	"github.com/ZecreyGaming/BlockChainWar/model"
	"github.com/topfreegames/pitaya/constants"
	"github.com/topfreegames/pitaya/v2"
	"github.com/topfreegames/pitaya/v2/component"
	"go.uber.org/zap"
)

type Room struct {
	component.Base
	app       pitaya.Pitaya
	cfg       *config.Config
	db        *db.Client
	sdkClient *sdk.Client
	game      *game.Game
}

func RegistRoom(app pitaya.Pitaya, db *db.Client, cfg *config.Config, game *game.Game) {
	err := app.GroupCreate(context.Background(), config.ChatRoomName)
	if err != nil {
		panic(err)
	}

	app.Register(&Room{
		app:  app,
		db:   db,
		cfg:  cfg,
		game: game,
	},
		component.WithName(config.ChatRoomName),
		component.WithNameFunc(strings.ToLower),
	)
}

// JoinResponse represents the result of joining room
type JoinResponse struct {
	Code     int           `json:"code"`
	Result   string        `json:"result"`
	GameInfo game.GameInfo `json:"game_info"`
}

type MessageResponse struct {
	Code   int    `json:"code"`
	Result string `json:"result"`
}

// NewUser message will be received when new user join room
type NewUser struct {
	Content string `json:"content"`
}

// Join room 玩家进入游戏
func (r *Room) Join(ctx context.Context, player *model.Player) (*JoinResponse, error) {
	//fmt.Println(fmt.Sprintf("Join player%v", player))
	s := r.app.GetSessionFromCtx(ctx)
	fakeUID := s.ID()                              // just use s.ID as uid !!!
	err := s.Bind(ctx, strconv.Itoa(int(fakeUID))) // binding session uid

	if err != nil && err != constants.ErrSessionAlreadyBound {
		return nil, pitaya.Error(err, "RH-000", map[string]string{"failed": "bind"})
	}
	//delete .zec suffix use zecrey nft sdk not need add .zec suffix will add it automatic
	name := strings.TrimSuffix(player.Name, ".zec")
	//set playerPk
	playerInfo, err := sdk.GetAccountInfo(name)
	if err != nil {
		return nil, pitaya.Error(err, "RH-500", map[string]string{"failed": "GetAccountInfo fail", "error": err.Error()})
	}
	player.L2publicKey = playerInfo.Account.AccountPk
	// offset, limit := 0, 100
	// // get last 30 messages
	// messages, err := r.db.Message.ListLatest(offset, limit)
	// if err != nil {
	// 	return nil, pitaya.Error(err, "RH-500", map[string]string{"failed": "get messages"})
	// }
	// s.Push("onHistoryMessage", messages)

	if err := r.db.Player.Create(player); err != nil {
		zap.L().Error("create player failed", zap.Error(err))
		return nil, pitaya.Error(err, "RH-500", map[string]string{"failed": "create player, db issue"})
	}

	// new user join group
	r.app.GroupAddMember(ctx, config.ChatRoomName, s.UID()) // add session to group

	// on session close, remove it from group
	s.OnClose(func() {
		r.app.GroupRemoveMember(ctx, config.ChatRoomName, s.UID())
	})

	info, err := r.game.GetGameInfo()
	if err != nil {
		return nil, pitaya.Error(err, "RH-500", map[string]string{"failed": "get game info", "error": err.Error()})
	}

	return &JoinResponse{Result: "success", GameInfo: info}, nil
}

// Message sync last message to all members
func (r *Room) Message(ctx context.Context, msg *model.Message) (*MessageResponse, error) {
	//fmt.Println("msg:", msg.Message)
	err := r.db.Message.Create(msg)
	if err != nil {
		zap.L().Error("save message failed", zap.Error(err))
	}
	player, err := r.db.Player.Get(msg.PlayerID)
	if err != nil {
		zap.L().Error("player not join game can`t send message", zap.Error(err))
		return nil, pitaya.Error(err, "RH-400", map[string]string{"failed": fmt.Sprintf("player(%d) not join game can`t send message err=%s", msg.Player.PlayerID, err)})
	}
	b, err := sdk.VerifyMessage(player.L2publicKey, msg.SignedMessage, msg.Message)
	if err != nil {
		zap.L().Error("sdk.VerifyMessage err failed", zap.Error(err))
		return nil, pitaya.Error(err, "RH-400", map[string]string{"failed": fmt.Sprintf("sdk.VerifyMessage failed err:%s", err)})
	}
	if !b {
		zap.L().Error("sdk.VerifyMessage  failed", zap.Error(err))
		return nil, pitaya.Error(err, "RH-400", map[string]string{"failed": fmt.Sprintf("sdk.VerifyMessage  failed")})
	}
	p, err := r.db.Player.Get(msg.PlayerID)
	if err != nil {
		zap.L().Error("get player failed", zap.Error(err))
		return nil, pitaya.Error(err, "RH-400", map[string]string{"failed": "get player, playerID not found"})
	}

	msg.Player = p
	err = r.app.GroupBroadcast(ctx, r.cfg.FrontendType, config.ChatRoomName, "onMessage", msg)
	if err != nil {
		zap.L().Error("broadcast message failed", zap.Error(err))
	}

	if camp := game.DecideCamp(msg.Message); camp != game.Empty && r.game != nil {
		r.game.StartRound(player.Name) //start by first people
		if err := r.db.Player.AddVote(&model.PlayerVote{
			GameID:   r.game.GetGameID(),
			PlayerID: msg.PlayerID,
			Camp:     uint8(camp),
		}); err == nil {
			r.app.GroupBroadcast(ctx, r.cfg.FrontendType, config.GameRoomName, "onPlayerJoin", p)
			r.game.AddPlayer(msg.PlayerID, game.DecideCamp(msg.Message))
		} else {
			zap.L().Error("add player vote failed", zap.Error(err))
		}
	}
	return &MessageResponse{
		Result: "success",
	}, nil
}
