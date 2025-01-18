package config

import (
	"fmt"
	"github.com/pelletier/go-toml/v2"
	"os"
)

type Config struct {
	DBServers []DbServer `toml:"db_servers"`
	Rules     []Rule     `toml:"rules"`
	Schedules []Schedule `toml:"scheduler"`
}

type DbServer struct {
	Name          string `toml:"Name"`
	Type          string `toml:"Type"`
	ConnStringVar string `toml:"ConnStringVar"`
}

type Schedule struct {
	Server  string `toml:"Server"`
	Rule    string `toml:"Rule"`
	CronStr string `toml:"CronStr"`
}

type Rule struct {
	Name        string `toml:"Name"`
	Description string `toml:"Description"`
	DbType      string `toml:"DbType"`
	Query       string `toml:"Query"`
}

func LoadConfigBytes(data []byte) (Config, error) {
	var payload struct {
		DBServers []DbServer `toml:"db_servers"`
		Rules     []Rule     `toml:"rules"`
		Schedules []Schedule `toml:"schedules"`
	}
	if err := toml.Unmarshal(data, &payload); err != nil {
		return Config{}, fmt.Errorf("error during Config Unmarshal(): %w", err)
	}

	config := Config{
		DBServers: payload.DBServers,
		Rules:     payload.Rules,
		Schedules: payload.Schedules,
	}
	return config, nil
}

func (server DbServer) GetConnString() (string, error) {
	connStr := os.Getenv(server.ConnStringVar)
	if connStr == "" {
		return "", fmt.Errorf("environment variable %s not found or empty", server.ConnStringVar)
	}
	return connStr, nil
}
