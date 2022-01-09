package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/open-policy-agent/opa/plugins"
)

const PluginName = "pingpong_plugin"

const DefaultPort = 9999
const DefaultMessage = "Pong!"

type Config struct {
	Port int32  `json:"port"`
	Msg  string `json:"msg"`
}

type PingPongServer struct {
	manager *plugins.Manager
	mtx     sync.Mutex
	config  Config
}

func (p *PingPongServer) Start(ctx context.Context) error {
	log.Printf("Start PingPong Server, config=%+v\n", p.config)

	start(&p.config)

	p.manager.UpdatePluginStatus(PluginName, &plugins.Status{State: plugins.StateOK})

	return nil
}

func (p *PingPongServer) Stop(ctx context.Context) {
	log.Println("Stop PingPong Server")

	p.manager.UpdatePluginStatus(PluginName, &plugins.Status{State: plugins.StateNotReady})
}

func (p *PingPongServer) Reconfigure(ctx context.Context, config interface{}) {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	p.config = config.(Config)
}

func start(config *Config) {
	listen := fmt.Sprintf(":%d", config.Port)

	log.Printf("listen addr=%s", listen)

	http.HandleFunc("/ping", func(res http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(res, "%s\n", config.Msg)
	})
	log.Fatal(http.ListenAndServe(listen, nil))
}
