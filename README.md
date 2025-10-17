# PG-Porter

## Overview

PG-Porter is a high-performance, command-line tool for exporting large datasets from a PostgreSQL database to a CSV file. It leverages PostgreSQL's efficient `COPY` command to ensure fast data transfer.

### Features

- **High-Speed Export:** Utilizes the native `COPY` command for maximum performance.
- **Flexible Configuration:** Configure database connections via a `.env` file or directly with psql-style command-line flags.
- **User-Friendly:** Displays a loading indicator during export operations so you know it's working.
- **Standalone Tool:** Compiles into a single binary with no external dependencies required at runtime.

## Prerequisites

- **Go:** Version 1.24 or higher.
- **PostgreSQL:** A running instance of PostgreSQL.

## Installation

Clone the repository and build the application:

```bash
git clone https://github.com/your-repo/pg-porter.git
cd pg-porter
go build -o pg-porter main.go
```

## Configuration

You can configure PG-Porter in two ways:

1.  **Using a `.env` file:**
    Create a `.env` file in the root of the project. This method is recommended for development.
    ```
    DB_HOST=localhost
    DB_USER=your_user
    DB_PASS=your_password
    DB_PORT=5432
    DB_NAME=pgbench_db
    DB_SSLMODE=disable
    ```

2.  **Using Command-Line Flags:**
    You can provide connection details directly as command-line arguments. This is useful for production environments or scripting.

**Configuration Precedence:** Command-line flags will always override settings in the `.env` file.

## Usage

Execute the `pg-porter` binary with the desired flags.

**Example:**

```bash
./pg-porter --sql="SELECT aid, bid, abalance FROM pgbench_accounts ORDER BY aid ASC" --out="output.csv" -U myuser -d mydb
```

While the export is in progress, you will see a loading spinner.

### Available Parameters

| Flag      | Environment Variable | Description                                           | Default     |
|-----------|----------------------|-------------------------------------------------------|-------------|
| `-sql`    |                      | **Required.** The SQL query to export.                |             |
| `-out`    |                      | **Required.** The output path for the CSV file.       |             |
| `-U`      | `DB_USER`            | **Required.** The database user name.                 |             |
| `-d`      | `DB_NAME`            | **Required.** The database name.                      |             |
| `-H`      | `DB_HOST`            | The database server host.                             | `localhost` |
| `-p`      | `DB_PORT`            | The database server port.                             | `5432`      |
| `-W`      | `DB_PASS`            | The database user's password.                         |             |
| `-sslmode`| `DB_SSLMODE`         | The SSL mode for the connection.                      | `prefer`    |
| `-timeout`|                      | The connection timeout in seconds.                    | `180`       |
| `-dsn`    |                      | A full DSN string (overrides all other DB flags).     |             |


## Sample Result

Running the application will produce output similar to the following:

```
$ ./pg-porter --sql="SELECT aid, bid, abalance FROM pgbench_accounts ORDER BY aid ASC" --out="output.csv" -U pguser -d pgbench_db
sql	: SELECT aid, bid, abalance FROM pgbench_accounts ORDER BY aid ASC
output	: output.csv
rows	: 1000000
elapsed	: 4.21 seconds

$ tail output.csv
999991,10,0
999992,10,0
999993,10,0
999994,10,0
999995,10,0
999996,10,0
999997,10,0
999998,10,0
999999,10,0
1000000,10,0
```

## Performance Testing

To simulate a large dataset for performance testing, you can use `pgbench` to generate sample data in your PostgreSQL database. The following command will generate approximately 1 million rows.

```bash
pgbench -i -s 10 pgbench_db
```
