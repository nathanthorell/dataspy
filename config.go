package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/pelletier/go-toml/v2"
)

func LoadConfig(filename string) ([]Connection, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error when opening file: %w", err)
	}
	var payload struct {
		Servers []Connection `json:"servers"`
	}
	if err := json.Unmarshal(content, &payload); err != nil {
		return nil, fmt.Errorf("error during Config Unmarshal(): %w", err)
	}
	return payload.Servers, nil
}

func LoadRules(filename string) ([]Rule, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error when opening file: %w", err)
	}

	var payload struct {
		Rules []Rule `toml:"rules"`
	}
	if err := toml.Unmarshal(content, &payload); err != nil {
		return nil, fmt.Errorf("error during Rules Unmarshal(): %w", err)
	}
	return payload.Rules, nil
}

func MapServerRules(connections []Connection, rules []Rule) map[string]ServerRules {
	serverRulesMap := make(map[string]ServerRules)

	for _, connection := range connections {
		serverRules := ServerRules{
			Server: connection,
			Rules:  make([]Rule, 0),
		}
		serverRulesMap[connection.Type] = serverRules
	}

	for _, rule := range rules {
		if serverRules, ok := serverRulesMap[rule.DbType]; ok {
			serverRules.Rules = append(serverRules.Rules, rule)
			serverRulesMap[rule.DbType] = serverRules
		}
	}

	return serverRulesMap
}
