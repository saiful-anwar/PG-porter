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
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/joho/godotenv"
)

// Config holds all configuration for the application.
type Config struct {
	SQL      string
	Out      string
	Timeout  int
	DSN      string
	User     string
	DbName   string
	Host     string
	Port     string
	Password string
	SslMode  string
}

// NewConfig creates a new Config object from command-line flags and environment variables.
func NewConfig(args []string) (*Config, error) {
	cfg := &Config{}
	fs := flag.NewFlagSet("pg-porter", flag.ContinueOnError)

	// Define flags
	fs.StringVar(&cfg.SQL, "sql", "", "SQL Command")
	fs.StringVar(&cfg.Out, "out", "", "Output file path")
	fs.StringVar(&cfg.DSN, "dsn", "", "Database connection string")
	fs.StringVar(&cfg.User, "U", "", "Database user")
	fs.StringVar(&cfg.DbName, "d", "", "Database name")
	fs.StringVar(&cfg.Host, "H", "", "Database host")
	fs.StringVar(&cfg.Port, "p", "", "Database port")
	fs.StringVar(&cfg.Password, "W", "", "Database password")
	fs.StringVar(&cfg.SslMode, "sslmode", "", "SSL mode (disable, allow, prefer, require, verify-ca, verify-full)")

	timeoutStr := fs.String("timeout", "180", "Connection Timeout")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	// Parse timeout
	timeout, err := strconv.Atoi(*timeoutStr)
	if err != nil {
		return nil, fmt.Errorf("error converting timeout to int: %w", err)
	}
	cfg.Timeout = timeout

	// Load .env file if it exists
	godotenv.Load()

	// Get values from flags, then env, then default
	cfg.Host = getFlagOrEnv(cfg.Host, "DB_HOST", "localhost")
	cfg.User = getFlagOrEnv(cfg.User, "DB_USER", "")
	cfg.Password = getFlagOrEnv(cfg.Password, "DB_PASS", "")
	cfg.Port = getFlagOrEnv(cfg.Port, "DB_PORT", "5432")
	cfg.DbName = getFlagOrEnv(cfg.DbName, "DB_NAME", "")
	cfg.SslMode = getFlagOrEnv(cfg.SslMode, "DB_SSLMODE", "prefer")

	// Validate required fields
	if cfg.DSN == "" {
		if cfg.User == "" {
			return nil, fmt.Errorf("database user is required")
		}
		if cfg.DbName == "" {
			return nil, fmt.Errorf("database name is required")
		}
	}
	if cfg.SQL == "" {
		return nil, fmt.Errorf("sql command is required")
	}
	if cfg.Out == "" {
		return nil, fmt.Errorf("output file path is required")
	}

	return cfg, nil
}

// BuildDSN constructs the DSN string from the Config.
func (c *Config) BuildDSN() string {
	if c.DSN != "" {
		return c.DSN
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", c.User, url.QueryEscape(c.Password), c.Host, c.Port, c.DbName, c.SslMode)
}

func main() {
	cfg, err := NewConfig(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	duration := time.Duration(cfg.Timeout) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	pgDsn := cfg.BuildDSN()

	conn, err := pgx.Connect(ctx, pgDsn)
	if err != nil {
		log.Fatalf("Unable to init db connection: %v", err)
	}
	defer conn.Close(ctx)

	err = copyToCSV(context.Background(), conn, cfg.SQL, cfg.Out)
	if err != nil {
		log.Fatalf("unable to copy: %v", err)
	}
}

func getFlagOrEnv(flagValue, envKey, defaultValue string) string {
	if flagValue != "" {
		return flagValue
	}
	if envValue := os.Getenv(envKey); envValue != "" {
		return envValue
	}
	return defaultValue
}

func copyToCSV(ctx context.Context, conn *pgx.Conn, sql string, out string) error {
	start := time.Now()

	file, err := os.Create(out)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	type copyResult struct {
		res pgconn.CommandTag
		err error
	}
	done := make(chan copyResult)

	go func() {
		command := fmt.Sprintf(`COPY (%s) TO STDOUT WITH (FORMAT csv, HEADER, DELIMITER ',')`, sql)
		res, err := conn.PgConn().CopyTo(ctx, file, command)
		done <- copyResult{res, err}
	}()

	spinner := []string{"|", "/", "-", "\\"}
	i := 0
	for {
		select {
		case result := <-done:
			if result.err != nil {
				return fmt.Errorf("failed to copy: %w", result.err)
			}
			fmt.Print("\r") // Clear spinner

			end := time.Now()
			duration := end.Sub(start)
			fmt.Println("sql\t:", sql)
			fmt.Println("output\t:", out)
			fmt.Println("rows\t:", result.res.RowsAffected())
			fmt.Printf("elapsed\t: %.2f seconds\n", duration.Seconds())

			return nil
		default:
			fmt.Printf("\rProcessing... %s ", spinner[i])
			i = (i + 1) % len(spinner)
			time.Sleep(100 * time.Millisecond)
		}
	}
}
