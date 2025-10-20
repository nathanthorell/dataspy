package main

import (
	_ "embed"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/microsoft/go-mssqldb"

	"github.com/nathanthorell/dataspy/cmd"
)

//go:embed config/config.toml
var configTOML []byte

func main() {
	cmd.Execute(configTOML)
}
