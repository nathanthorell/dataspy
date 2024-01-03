package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/microsoft/go-mssqldb"
)

func ExecuteQuery(conn Connection, query string) {
	var connStr string
	switch conn.Type {
	case "mysql":
		connStr = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
			conn.User, conn.Password, conn.Server, conn.Port, conn.Database)
		fmt.Println("MySQL Connection String:", connStr)

	case "postgres":
		connStr = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			conn.Server, conn.Port, conn.User, conn.Password, conn.Database)
	case "mssql":
		connStr = fmt.Sprintf("server=%s;port=%s;user id=%s;password=%s;database=%s;",
			conn.Server, conn.Port, conn.User, conn.Password, conn.Database)
	default:
		log.Fatalf("unknown connection type: %s", conn.Type)
	}

	// Connect to database
	db, err := sql.Open(conn.Type, connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Check if the connection is successful
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	// Execute a simple SQL query
	var result string
	err = db.QueryRow(query).Scan(&result)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result)
}
