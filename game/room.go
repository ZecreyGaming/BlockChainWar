package game

import (
	"context"
	"github.com/ZecreyGaming/BlockChainWar/game/cronjob/zecreyface"
	"go.uber.org/zap"

	"strconv"
	"strings"
	"time"

	"github.com/ZecreyGaming/BlockChainWar/config"
	"github.com/ZecreyGaming/BlockChainWar/db"
	"github.com/ZecreyGaming/BlockChainWar/model"
	"github.com/topfreegames/pitaya/constants"
	"github.com/topfreegames/pitaya/v2"
	"github.com/topfreegames/pitaya/v2/component"
)

type Room struct {
	component.Base
	ctx context.Context
	app pitaya.Pitaya
	cfg *config.Config
	db  *db.Client

	tickerCancel context.CancelFunc
	game         *Game
}

type GameUpdate struct {
	Data []byte `json:"data"`
}

func RegistRoom(app pitaya.Pitaya, db *db.Client, cfg *config.Config, sdkClient *zecreyface.Client) *Game {
	err := app.GroupCreate(context.Background(), config.GameRoomName)
	if err != nil {
		panic(err)
	}
	r := &Room{
		app: app,
		db:  db,
		cfg: cfg,
	}
	r.ctx, r.tickerCancel = context.WithCancel(context.Background())
	r.game = NewGame(r.ctx, cfg, db, sdkClient, r.onGameStart, r.onGameStop, r.onCampVotesChange)
	app.Register(r,
		component.WithName(config.GameRoomName),
		component.WithNameFunc(strings.ToLower),
	)
	return r.game
}

func (r *Room) AfterInit() {
	stateChan := r.game.start()
	go func() {
		// ticker := time.Tick(time.Duration(33) * time.Millisecond)
		ticker := time.NewTicker(time.Duration(1000/r.cfg.FPS) * time.Millisecond).C
		for {
			select {
			//case nextRoundChan := <-r.game.stopSignalChan:
			//	<-nextRoundChan
			case <-r.ctx.Done():
				return
			default:
				s := <-stateChan
				<-ticker
				err := r.app.GroupBroadcast(context.Background(), r.cfg.FrontendType, config.GameRoomName, "onUpdate", GameUpdate{Data: s})
				if err != nil {
					zap.L().Error("broadcast onUpdate failed", zap.Error(err))
				}
			}
		}
	}()
}

func (r *Room) Shutdown() {
	r.tickerCancel()
}

// JoinResponse represents the result of joining room
type JoinResponse struct {
	Code       int    `json:"code"`
	Result     string `json:"result"`
	GameStatus uint8  `json:"game_status"`
	Winner     uint8  `json:"winner"`
}

// NewUser message will be received when new user join room
type NewUser struct {
	Content string `json:"content"`
}

// Join room
func (r *Room) Join(ctx context.Context, msg []byte) (*JoinResponse, error) {
	// if r.game == nil || r.game.GameStatus != GameRunning {
	// 	return nil, pitaya.Error(fmt.Errorf("GAME_NOT_START"), "GAME_NOT_START", map[string]string{"failed": "game not start"})
	// }

	s := r.app.GetSessionFromCtx(ctx)
	fakeUID := s.ID()                              // just use s.ID as uid !!!
	err := s.Bind(ctx, strconv.Itoa(int(fakeUID))) // binding session uid

	if err != nil && err != constants.ErrSessionAlreadyBound {
		return nil, pitaya.Error(err, "RH-000", map[string]string{"failed": "bind"})
	}

	// uids, err := r.app.GroupMembers(ctx, config.GameRoomName)
	// if err != nil {
	// 	return nil, err
	// }
	// s.Push("onMembers", &AllMembers{Members: uids})

	// new user join group
	r.app.GroupAddMember(ctx, config.GameRoomName, s.UID()) // add session to group
	//todo 发游戏状态
	// notify others
	r.onJoin(ctx, false)

	// on session close, remove it from group
	s.OnClose(func() {
		r.app.GroupRemoveMember(ctx, config.GameRoomName, s.UID())
	})
	gameInfo, err := r.game.GetGameInfo()
	//fmt.Println("gameInfo === ", JoinResponse{Result: "success", Code: 0, GameStatus: uint8(gameInfo.GameStatus), Winner: gameInfo.WinnerId})
	return &JoinResponse{Result: "success", Code: 0, GameStatus: uint8(gameInfo.GameStatus), Winner: gameInfo.WinnerId}, nil //code == 0 join game
}

// onJoin room
func (r *Room) onJoin(ctx context.Context, replay bool) {
	mi := MapInfo{
		Row:        mapRow,
		Column:     mapColumn,
		CellWidth:  cellWidth,
		CellHeight: cellHeight,

		Item:   AllItems,
		Replay: replay,
	}

	pids := []uint64{}
	r.game.Players.Range(func(key, value interface{}) bool {
		pids = append(pids, key.(uint64))
		return true
	})

	mi.Players, _ = r.db.Player.List(pids...)
	r.app.GroupBroadcast(ctx, r.cfg.FrontendType, config.GameRoomName, "onJoin", mi)
}

func (r *Room) onGameStart(ctx context.Context) {
	info, _ := r.game.GetGameInfo()
	r.app.GroupBroadcast(r.ctx, r.cfg.FrontendType, config.GameRoomName, "onGameStart", info)
	r.app.GroupBroadcast(r.ctx, r.cfg.FrontendType, config.ChatRoomName, "onGameStart", info)
	r.onJoin(ctx, true)
}

// getGameInfo room
func (r *Room) getGameInfo(ctx context.Context, msg []byte) (*JoinResponse, error) {
	info, _ := r.game.GetGameInfo()
	r.app.GroupBroadcast(r.ctx, r.cfg.FrontendType, config.ChatRoomName, "getGameInfo", info)
	r.onJoin(ctx, true)
	return &JoinResponse{Result: "success", Code: 0}, nil //code == 0 join game
}

func (r *Room) onUpdate(s []byte) {
	r.app.GroupBroadcast(r.ctx, r.cfg.FrontendType, config.GameRoomName, "onUpdate", GameUpdate{Data: s})
}

func (r *Room) onGameStop(ctx context.Context) {
	stop := r.game.GetGameStop()
	//fmt.Println("winner info ", stop)
	r.app.GroupBroadcast(ctx, r.cfg.FrontendType, config.GameRoomName, "onGameStop", stop)
	r.app.GroupBroadcast(ctx, r.cfg.FrontendType, config.ChatRoomName, "onGameStop", stop)
}

func (r *Room) onCampVotesChange(camp Camp, votes int32) {
	r.app.GroupBroadcast(r.ctx, r.cfg.FrontendType, config.ChatRoomName, "onCampVotesChange", CampVotesChange{
		Camp:  camp,
		Votes: votes,
	})
}

// TODO
type MapInfo struct {
	Row    uint32 `json:"row"`
	Column uint32 `json:"column"`

	CellWidth  uint32 `json:"cell_width"`
	CellHeight uint32 `json:"cell_height"`

	Item    []Item         `json:"items"`
	Players []model.Player `json:"players"`
	Replay  bool           `json:"replay"`
}

type CampVotesChange struct {
	Camp  Camp  `json:"camp"`
	Votes int32 `json:"votes"`
}
