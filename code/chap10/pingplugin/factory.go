package main

import (
	"log"

	"github.com/open-policy-agent/opa/plugins"
	"github.com/open-policy-agent/opa/util"
)

type Factory struct{}

func (Factory) New(m *plugins.Manager, config interface{}) plugins.Plugin {

	log.Printf("Config=%+v\n", config)

	m.UpdatePluginStatus(PluginName, &plugins.Status{State: plugins.StateNotReady})

	return &PingPongServer{
		manager: m,
		config:  config.(Config),
	}
}

func (Factory) Validate(_ *plugins.Manager, config []byte) (interface{}, error) {
	parsedConfig := Config{}
	err := util.Unmarshal(config, &parsedConfig)

	if err != nil {
		log.Println("Error occured while validate config:%v\n", err)
	} else {
		if parsedConfig.Msg == "" {
			parsedConfig.Msg = DefaultMessage
		}

		if parsedConfig.Port == 0 {
			parsedConfig.Port = DefaultPort
		}
	}

	return parsedConfig, err
}
