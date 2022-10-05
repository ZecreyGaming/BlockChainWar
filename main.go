package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/COAOX/zecrey_warrior/model"
	"github.com/topfreegames/pitaya/v2"
	"github.com/topfreegames/pitaya/v2/acceptor"
	"github.com/topfreegames/pitaya/v2/component"
	"github.com/topfreegames/pitaya/v2/config"
	"github.com/topfreegames/pitaya/v2/groups"
	"github.com/topfreegames/pitaya/v2/timer"
)

type Room struct {
	component.Base
	timer *timer.Timer
	app   pitaya.Pitaya

	timerCancel context.CancelFunc
	game        *model.Game
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

// AllMembers contains all members uid
type AllMembers struct {
	Members []string `json:"members"`
}

// JoinResponse represents the result of joining room
type JoinResponse struct {
	Code   int    `json:"code"`
	Result string `json:"result"`
}

func NewRoom(app pitaya.Pitaya) *Room {
	return &Room{
		app:  app,
		game: model.NewGame(),
	}
}

func (r *Room) AfterInit() {
	ctx, cancel := context.WithCancel(context.Background())
	r.timerCancel = cancel
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
		ticker := time.Tick(time.Second)
		for {
			select {
			case s := <-stateChan:
				<-ticker
				r.app.GroupBroadcast(context.Background(), "zecrey_warrior", "room", "onUpdate", s)
			case <-ctx.Done():
				return
			}
		}
	}()
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
		fmt.Println("error broadcasting message", err)
	}
}

var app pitaya.Pitaya

func main() {
	conf := configApp()
	builder := pitaya.NewDefaultBuilder(true, "zecrey_warrior", pitaya.Standalone, map[string]string{}, *conf)
	builder.AddAcceptor(acceptor.NewWSAcceptor(":3250"))
	builder.Groups = groups.NewMemoryGroupService(*config.NewDefaultMemoryGroupConfig())
	builder.Serializer = model.NewSerializer()
	app = builder.Build()

	defer app.Shutdown()

	err := app.GroupCreate(context.Background(), "room")
	if err != nil {
		panic(err)
	}

	// rewrite component and handler name
	room := NewRoom(app)
	app.Register(room,
		component.WithName("room"),
		component.WithNameFunc(strings.ToLower),
	)

	log.SetFlags(log.LstdFlags | log.Llongfile)

	http.Handle("/web/", http.StripPrefix("/web/", http.FileServer(http.Dir("web"))))

	go http.ListenAndServe(":3251", nil)

	app.Start()
}

func configApp() *config.BuilderConfig {
	conf := config.NewDefaultBuilderConfig()
	conf.Pitaya.Heartbeat.Interval = time.Duration(3 * time.Second)
	conf.Pitaya.Buffer.Agent.Messages = 32
	conf.Pitaya.Handler.Messages.Compression = false
	conf.Metrics.Prometheus.Enabled = true
	return conf
}
