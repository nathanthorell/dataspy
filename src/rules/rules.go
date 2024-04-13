package rules

import (
    "github.com/nathanthorell/dataspy/db"
)

type Rule struct {
	Name           string `toml:"Name"`
	Description    string `toml:"Description"`
	DbType         string `toml:"DbType"`
	Query          string `toml:"Query"`
}

type ServerRules struct {
	Server db.Connection
	Rules  []Rule
}

func (r Rule) FilterValue() string {
	return r.Name
}

func MapServerRules(connections []db.Connection, rules []Rule) map[string]ServerRules {
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