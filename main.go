package main

import (
	"fmt"
	"log"
)

type ServerRules struct {
	Server Connection
	Rules  []Rule
}

func main() {
	// Load Server Config
	servers, err := LoadConfig("config.json")
	if err != nil {
		log.Fatal(err)
	}

	// Load Rules Config
	rules, err := LoadRules("rules.toml")
	if err != nil {
		log.Fatal(err)
	}

	serverRulesMap := MapServerRules(servers, rules)

	// Execute Querys
	for _, serverRules := range serverRulesMap {
		for _, rule := range serverRules.Rules {
			fmt.Println("---------------------------------------")
			fmt.Println("Running Rule: ", rule.Name)
			ExecuteQuery(serverRules.Server, rule.Query)
			fmt.Println("---------------------------------------")
			fmt.Printf("\n")
		}
	}
}
