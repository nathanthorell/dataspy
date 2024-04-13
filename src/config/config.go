package config

import (
	"encoding/json"
	"fmt"

	"github.com/nathanthorell/dataspy/db"
	"github.com/nathanthorell/dataspy/rules"
	"github.com/pelletier/go-toml/v2"
)

func LoadConfigBytes(data []byte) ([]db.Connection, error) {
	var payload struct {
		Servers []db.Connection `json:"servers"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("error during Config Unmarshal(): %w", err)
	}
	return payload.Servers, nil
}

func LoadRulesBytes(data []byte) ([]rules.Rule, error) {
	var payload struct {
		Rules []rules.Rule `toml:"rules"`
	}
	if err := toml.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("error during Rules Unmarshal(): %w", err)
	}
	return payload.Rules, nil
}
