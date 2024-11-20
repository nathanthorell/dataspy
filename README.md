# dataspy

A simple SQL query runner command line tool for checking SQL "rules" against sets of databases.
Currently supports PostgreSQL, SQL Server, and mySql.

## Configuration

- Replace `./config/config.json` with your database configuration
- Replace `./config/rules.toml` with your query rules

## Build and Execute

- `go build -o ./build/dataspy ./.`
- `./build/dataspy`

## Docker

- `docker build -t dataspy .`
- `docker run --network="host" dataspy`
