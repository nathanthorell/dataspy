package main

import (
	_ "embed"
	"fmt"
	"log"

	"github.com/nathanthorell/dataspy/config"
	"github.com/nathanthorell/dataspy/db"
	"github.com/nathanthorell/dataspy/rules"
)

//go:embed config/config.json
var configJSON []byte

//go:embed config/rules.toml
var rulesTOML []byte

func main() {
	servers, err := config.LoadConfigBytes(configJSON)
	if err != nil {
		log.Fatal(err)
	}

	rules_conf, err := config.LoadRulesBytes(rulesTOML)
	if err != nil {
		log.Fatal(err)
	}

	serverRulesMap := rules.MapServerRules(servers, rules_conf)

	// Execute Querys
	for _, serverRules := range serverRulesMap {
		for _, rule := range serverRules.Rules {
			fmt.Println("---------------------------------------")
			fmt.Println("Running Rule: ", rule.Name)
			db.ExecuteQuery(serverRules.Server, rule.Query)
			fmt.Println("---------------------------------------")
			fmt.Printf("\n")
		}
	}
}
