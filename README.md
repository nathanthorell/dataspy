# dataspy

dataspy is a lightweight database monitoring tool that quietly observes your databases for business rule violations. It supports running SQL-based rules against multiple database types and can schedule regular checks.

## Features

- **Multi-Database Support**: Currently supports PostgreSQL, MySQL, and SQL Server
- **Configurable Rules**: Define SQL queries to check for business rule violations
- **Scheduled Monitoring**: Run rules on configurable cron schedules
- **Environment-Based Configuration**: Secure connection string management via environment variables
- **Cross-Database Monitoring**: Run the same business rules against multiple databases or environments

## Configuration

### Project Structure

```none
├── /
│── main.go              # Application entry point
├── config/              # Configuration management
│   └── config.toml      # Application configuration
├── db/                  # Database interaction
├── runner/              # Task scheduling and execution
└── Dockerfile           # Container build instructions
```

### Environment Setup

Create a `.env` file in your project directory with your database connection strings:

```env
PG_DBCONN="host=localhost port=5432 user=postgres password=secret dbname=postgres sslmode=disable"
MSSQL_DBCONN="server=localhost;user id=sa;password=secret;port=1433;database=master"
MySQL_DBCONN="root:secret@tcp(localhost:3306)/mysql"
```

### Application Configuration

Configuration is managed through a single TOML file (`config.toml`) with three main sections:

1. **Database Servers**

    ```toml
    [[db_servers]]
    Name = "Local Postgres"
    Type = "postgres"
    ConnStringVar = "PG_DBCONN"
    ```

1. **Rules**

    ```toml
    [[rules]]
    Name = "Check Data Consistency"
    Description = "Verify data integrity"
    DbType = "postgres"
    Query = """SELECT count(*) FROM table WHERE condition;"""
    ```

1. **Schedules**

    The cron format with seconds is:
    seconds minute hour day-of-month month day-of-week

    ```toml
    [[schedules]]
    Server = "Local Postgres"
    Rule = "Check Data Consistency"
    CronStr = "*/5 * * * *"  # Run every 5 minutes (at 0 seconds)
    ```

## Building and Running

```bash
# Build the application
go build -o ./build/dataspy .

# Run from project root (where your .env file is located)
./build/dataspy
```

### Docker

Alternatively run this with docker

```bash
# Build the image
docker build -t dataspy .

# Run example with environment variables
docker run --network="host" \
  -e PG_DBCONN="host=localhost port=5432 user=postgres password=secret dbname=postgres" \
  -e MSSQL_DBCONN="server=localhost;user id=sa;password=secret;port=1433;database=master" \
  -e MySQL_DBCONN="root:secret@tcp(localhost:3306)/mysql" \
  dataspy
```

## Contributing

This is an active work in progress. Contributions and suggestions are welcome.
