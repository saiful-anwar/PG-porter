package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func main() {
	// parse params
	sql := flag.String("sql", "", "SQL Command")
	out := flag.String("out", "", "Output file path")
	timeout := flag.String("timeout", "", "Connection Timeout")
	dsn := flag.String("dsn", "", "Database connection")
	flag.Parse()

	var connTimeout int = 180
	if *timeout != "" {
		t, err := strconv.Atoi(*timeout)
		if err != nil {
			log.Fatalf("Error converting string to int: %v\n", err)
		}

		connTimeout = t
	}

	duration := time.Duration(connTimeout) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	var pgDsn string
	if *dsn != "" {
		pgDsn = *dsn
	} else {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}

		// read env
		host := os.Getenv("DB_HOST")
		user := os.Getenv("DB_USER")
		pass := url.QueryEscape(os.Getenv("DB_PASS"))
		port := os.Getenv("DB_PORT")
		db := os.Getenv("DB_NAME")

		pgDsn = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, pass, host, port, db)
	}

	conn, err := pgx.Connect(ctx, pgDsn)
	if err != nil {
		log.Fatalf("Unable to init db connection: %v", err)
	}

	defer conn.Close(ctx)

	// execute copy command
	err = copyToCSV(context.Background(), conn, *sql, *out)
	if err != nil {
		log.Fatalf("unable to copy: %v", err)
	}
}

func copyToCSV(ctx context.Context, conn *pgx.Conn, sql string, out string) error {
	start := time.Now()

	// Create or open the CSV file
	file, err := os.Create(out)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush() // Ensure data is written to file

	command := fmt.Sprintf(`COPY (%s) TO STDOUT WITH (FORMAT csv, HEADER, DELIMITER ',')`, sql)
	res, err := conn.PgConn().CopyTo(ctx, file, command)
	if err != nil {
		return fmt.Errorf("failed to copy: %w", err)
	}

	end := time.Now()
	duration := end.Sub(start)
	fmt.Println("sql\t:", sql)
	fmt.Println("output\t:", out)
	fmt.Println("rows\t:", res.RowsAffected())
	fmt.Printf("elapsed\t: %.2f seconds\n", duration.Seconds())

	return nil
}
