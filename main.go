package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/COAOX/zecrey_warrior/chat"
	cfg "github.com/COAOX/zecrey_warrior/config"
	"github.com/COAOX/zecrey_warrior/db"
	"github.com/COAOX/zecrey_warrior/game"
	"github.com/topfreegames/pitaya/v2"
	"github.com/topfreegames/pitaya/v2/acceptor"
	"github.com/topfreegames/pitaya/v2/config"
	"github.com/topfreegames/pitaya/v2/groups"
)

var (
	configPath = flag.String("config", "./config/local.json", "Path to config file")
)

func main() {
	flag.Parse()
	cfg := cfg.Read(*configPath)

	builder := pitaya.NewDefaultBuilder(true, cfg.FrontendType, pitaya.Standalone, map[string]string{}, configApp())
	builder.AddAcceptor(acceptor.NewWSAcceptor(":3250"))
	builder.Groups = groups.NewMemoryGroupService(*config.NewDefaultMemoryGroupConfig())
	builder.Serializer = game.NewSerializer()
	app := builder.Build()

	defer app.Shutdown()

	database := db.NewClient(cfg.Database)

	// register game and chat
	game.RegistRoom(app, database, cfg)
	chat.RegistRoom(app, database, cfg)

	log.SetFlags(log.LstdFlags | log.Llongfile)

	http.Handle("/web/", http.StripPrefix("/web/", http.FileServer(http.Dir("web"))))

	go http.ListenAndServe(":3251", nil)

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
