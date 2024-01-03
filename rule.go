package main

type Rule struct {
	Name           string `toml:"Name"`
	Description    string `toml:"Description"`
	DbType         string `toml:"DbType"`
	Query          string `toml:"Query"`
}

func (r Rule) FilterValue() string {
	return r.Name
}
