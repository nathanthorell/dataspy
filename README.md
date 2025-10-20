# dataspy

dataspy is a lightweight database monitoring tool that checks your databases for business rule violations that you define. It supports running SQL-based rules against multiple database types and can schedule regular checks.

## Features

- **Multi-Database Support**: Currently supports PostgreSQL, MySQL, and SQL Server
- **Configurable Rules**: Define SQL queries to check for business rule violations
- **Scheduled Monitoring**: Run rules on configurable cron schedules
- **Environment-Based Configuration**: Secure connection string management via environment variables
- **Cross-Database Monitoring**: Run the same business rules against multiple databases or environments

## Quick Start

1. **Copy example files**:

   ```bash
   cp .env.example .env
   cp config/config.toml.example config/config.toml
   ```

2. **Edit `.env`** with your database credentials

3. **Edit `config/config.toml`** with your rules and schedules

4. **Run on-demand** or as a **daemon**:

   ```bash
   # Run a specific rule once
   ./dataspy run --rule "Get Postgres Version"

   # Run all rules once
   ./dataspy run --all

   # Start daemon mode (scheduled monitoring)
   ./dataspy daemon
   ```

## Configuration

### Project Structure

```none
├── /
│── main.go                    # Application entry point
├── cmd/                       # CLI commands
│   ├── root.go                # Root command setup
│   ├── daemon.go              # Daemon mode (scheduled)
│   └── run.go                 # On-demand execution
├── config/                    # Configuration management
│   ├── config.toml            # Application configuration (gitignored)
│   └── config.toml.example    # Example configuration
├── db/                        # Database interaction
├── runner/                    # Task scheduling and execution
├── storage/                   # BoltDB storage layer
├── logger/                    # Pretty logging
├── .env                       # Environment variables (gitignored)
├── .env.example               # Example environment file
└── Dockerfile                 # Container build instructions
```

### Environment Setup

Copy `.env.example` to `.env` and edit with your database connection strings:

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

# Run on-demand
./build/dataspy run --rule "Rule Name"
./build/dataspy run --all

# Run as daemon (scheduled monitoring)
./build/dataspy daemon

# Specify custom .env location
./build/dataspy --env /path/to/.env daemon
```

## CLI Commands

### `dataspy run`

Execute rules on-demand without waiting for scheduled execution.

**Flags:**

- `-r, --rule <name>` - Run a specific rule by name
- `-a, --all` - Run all configured rules

**Examples:**

```bash
# Run a single rule
dataspy run --rule "Get Postgres Version"

# Run all rules
dataspy run --all
```

### `dataspy daemon`

Start the scheduler to run rules on their configured cron schedules.

**Example:**

```bash
dataspy daemon
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
