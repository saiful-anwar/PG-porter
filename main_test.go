package main

import (
	"os"
	"testing"
)

func TestGetFlagOrEnv(t *testing.T) {
	t.Run("it should return flag value when present", func(t *testing.T) {
		if val := getFlagOrEnv("flag_val", "ENV_KEY", "default"); val != "flag_val" {
			t.Errorf("Expected flag_val, got %s", val)
		}
	})

	t.Run("it should return env value when flag is empty", func(t *testing.T) {
		os.Setenv("ENV_KEY", "env_val")
		defer os.Unsetenv("ENV_KEY")
		if val := getFlagOrEnv("", "ENV_KEY", "default"); val != "env_val" {
			t.Errorf("Expected env_val, got %s", val)
		}
	})

	t.Run("it should return default value when flag and env are empty", func(t *testing.T) {
		if val := getFlagOrEnv("", "ENV_KEY", "default"); val != "default" {
			t.Errorf("Expected default, got %s", val)
		}
	})
}

func TestNewConfig(t *testing.T) {
	t.Run("it should parse flags correctly", func(t *testing.T) {
		args := []string{
			"-sql", "SELECT 1",
			"-out", "test.csv",
			"-U", "myuser",
			"-d", "mydb",
			"-H", "myhost",
			"-p", "5433",
			"-W", "mypass",
			"-sslmode", "require",
			"-timeout", "60",
		}
		cfg, err := NewConfig(args)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if cfg.SQL != "SELECT 1" {
			t.Errorf("Wrong SQL: expected 'SELECT 1', got '%s'", cfg.SQL)
		}
		if cfg.Out != "test.csv" {
			t.Errorf("Wrong Out: expected 'test.csv', got '%s'", cfg.Out)
		}
		if cfg.User != "myuser" {
			t.Errorf("Wrong User: expected 'myuser', got '%s'", cfg.User)
		}
		if cfg.DbName != "mydb" {
			t.Errorf("Wrong DbName: expected 'mydb', got '%s'", cfg.DbName)
		}
		if cfg.Host != "myhost" {
			t.Errorf("Wrong Host: expected 'myhost', got '%s'", cfg.Host)
		}
		if cfg.Port != "5433" {
			t.Errorf("Wrong Port: expected '5433', got '%s'", cfg.Port)
		}
		if cfg.Password != "mypass" {
			t.Errorf("Wrong Password: expected 'mypass', got '%s'", cfg.Password)
		}
		if cfg.SslMode != "require" {
			t.Errorf("Wrong SslMode: expected 'require', got '%s'", cfg.SslMode)
		}
		if cfg.Timeout != 60 {
			t.Errorf("Wrong Timeout: expected 60, got %d", cfg.Timeout)
		}
	})

	t.Run("it should use environment variables as fallback", func(t *testing.T) {
		os.Setenv("DB_USER", "envuser")
		os.Setenv("DB_NAME", "envdb")
		os.Setenv("DB_HOST", "envhost")
		os.Setenv("DB_PORT", "envport")
		os.Setenv("DB_PASS", "envpass")
		os.Setenv("DB_SSLMODE", "allow")
		defer func() {
			os.Unsetenv("DB_USER")
			os.Unsetenv("DB_NAME")
			os.Unsetenv("DB_HOST")
			os.Unsetenv("DB_PORT")
			os.Unsetenv("DB_PASS")
			os.Unsetenv("DB_SSLMODE")
		}()

		args := []string{"-sql", "SELECT 1", "-out", "test.csv"}
		cfg, err := NewConfig(args)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if cfg.User != "envuser" {
			t.Errorf("Expected user from env var, got '%s'", cfg.User)
		}
		if cfg.DbName != "envdb" {
			t.Errorf("Expected dbname from env var, got '%s'", cfg.DbName)
		}
		if cfg.Host != "envhost" {
			t.Errorf("Expected host from env var, got '%s'", cfg.Host)
		}
		if cfg.Port != "envport" {
			t.Errorf("Expected port from env var, got '%s'", cfg.Port)
		}
		if cfg.Password != "envpass" {
			t.Errorf("Expected password from env var, got '%s'", cfg.Password)
		}
		if cfg.SslMode != "allow" {
			t.Errorf("Expected sslmode from env var, got '%s'", cfg.SslMode)
		}
	})

	t.Run("it should let flags override environment variables", func(t *testing.T) {
		os.Setenv("DB_USER", "envuser")
		defer os.Unsetenv("DB_USER")

		args := []string{"-sql", "SELECT 1", "-out", "test.csv", "-U", "flaguser", "-d", "flagdb"}
		cfg, err := NewConfig(args)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if cfg.User != "flaguser" {
			t.Errorf("Expected user from flag, got '%s'", cfg.User)
		}
	})

	t.Run("it should return error for missing required fields", func(t *testing.T) {
		testCases := [][]string{
			{},
			{"-out", "file.csv", "-U", "user", "-d", "db"},
			{"-sql", "SELECT 1", "-U", "user", "-d", "db"},
			{"-sql", "SELECT 1", "-out", "file.csv", "-d", "db"},
			{"-sql", "SELECT 1", "-out", "file.csv", "-U", "user"},
		}

		for _, args := range testCases {
			_, err := NewConfig(args)
			if err == nil {
				t.Errorf("Expected error for args: %v", args)
			}
		}
	})

	t.Run("it should build DSN from parts", func(t *testing.T) {
		cfg := &Config{
			User:     "myuser",
			Password: "mypassword",
			Host:     "myhost",
			Port:     "5432",
			DbName:   "mydb",
			SslMode:  "disable",
		}
		expectedDSN := "postgres://myuser:mypassword@myhost:5432/mydb?sslmode=disable"
		if dsn := cfg.BuildDSN(); dsn != expectedDSN {
			t.Errorf("Wrong DSN: expected '%s', got '%s'", expectedDSN, dsn)
		}
	})

	t.Run("it should return DSN from config if present", func(t *testing.T) {
		dsnString := "postgres://myuser:mypassword@myhost:5432/mydb?sslmode=disable"
		cfg := &Config{
			DSN: dsnString,
		}
		if dsn := cfg.BuildDSN(); dsn != dsnString {
			t.Errorf("Wrong DSN: expected '%s', got '%s'", dsnString, dsn)
		}
	})
}
