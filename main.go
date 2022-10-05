package main

import (
	"context"
	"flag"
	"log"
	"strings"
	"time"

	cfg "github.com/COAOX/zecrey_warrior/config"
	"github.com/COAOX/zecrey_warrior/db"
	"github.com/COAOX/zecrey_warrior/game"
	"github.com/COAOX/zecrey_warrior/state"
	"github.com/topfreegames/pitaya/v2"
	"github.com/topfreegames/pitaya/v2/acceptor"
	"github.com/topfreegames/pitaya/v2/component"
	"github.com/topfreegames/pitaya/v2/config"
	"github.com/topfreegames/pitaya/v2/groups"
)

var (
	configPath = flag.String("config", "/etc/appconf/configs.json", "Path to config file")
)

func main() {
	flag.Parse()
	cfg := cfg.Read(*configPath)

	builder := pitaya.NewDefaultBuilder(true, "zecrey_warrior", pitaya.Standalone, map[string]string{}, configApp())
	builder.AddAcceptor(acceptor.NewWSAcceptor(":3250"))
	builder.Groups = groups.NewMemoryGroupService(*config.NewDefaultMemoryGroupConfig())
	builder.Serializer = state.NewSerializer()
	app := builder.Build()

	defer app.Shutdown()

	err := app.GroupCreate(context.Background(), "room")
	if err != nil {
		panic(err)
	}

	database := db.NewClient(cfg.Database)
	// rewrite component and handler name
	room := game.NewRoom(app, database, cfg)
	app.Register(room,
		component.WithName("room"),
		component.WithNameFunc(strings.ToLower),
	)

	log.SetFlags(log.LstdFlags | log.Llongfile)

	app.Start()
}

func configApp() config.BuilderConfig {
	conf := config.NewDefaultBuilderConfig()
	conf.Pitaya.Heartbeat.Interval = time.Duration(3 * time.Second)
	conf.Pitaya.Buffer.Agent.Messages = 32
	conf.Pitaya.Handler.Messages.Compression = false
	conf.Metrics.Prometheus.Enabled = true
	return *conf
}
