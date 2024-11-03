# EXPORT PostgreSQL data using COPY Command

## Overview

This project provides a method to export large amounts of data from a PostgreSQL database using the COPY command, reading the data and saving it to a CSV file.

## Prerequisites

- **Go** installed (version 1.16 or higher is recommended)
- **PostgreSQL** installed and running
- **pgbench** utility to generate sample data

## Generate Sample Data

To simulate a large dataset, use `pgbench` to generate approximately 1 million rows in your PostgreSQL database. Run the following command:

```bash
pgbench -i -s 10 pgbench_db
```

The -s 10 option sets the scaling factor, generating around 1 million rows. Adjust the scaling factor as needed based on available resources.

## Configuration
1. Copy the example environment file and rename it:
```
cp env.example .env
```
2. Update the .env file with your PostgreSQL database credentials:
```
DB_HOST=your_host
DB_USER=your_user
DB_PASS=your_password
DB_PORT=your_port
DB_NAME=pgbench_db
```

## How to Run the Application
To run the main program, build the Go application and execute it with the following commands:
```
➜ go build main.go
➜ /pg-copy --sql="SELECT aid, bid, abalance FROM pgbench_accounts WHERE ORDER BY aid ASC" --out="output.csv
```

## Available parameters
- --sql: The SQL command to be executed.
- --out: The file path where the generated CSV will be saved.
- --timeout: The timeout for query execution in seconds (default: 180 seconds).
- --dsn: Use custom dsn instead of reading connection variable from `.env` file

## Sample Test Result
Running the application with the provided SQL command will yield output similar to the following:

```
➜ ./pg-copy --sql="SELECT  aid, bid, abalance FROM pgbench_accounts WHERE ORDER BY aid ASC" --out="output.csv"
sql	: SELECT  aid, bid, abalance FROM pgbench_accounts ORDER BY aid ASC
output	: output.csv
rows	: 1000000
elapsed	: 4.21 seconds

➜ tail output.csv
999990,10,0
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