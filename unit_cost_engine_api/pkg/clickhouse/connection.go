package clickhouse

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	ch "github.com/ClickHouse/clickhouse-go/v2"
)

type Config struct {
	Addr     []string
	Database string
	Username string
	Password string
	Table    string
}

func ConfigFromEnv() (Config, error) {
	table := getEnv("CLICKHOUSE_TABLE", "cloud_costs")
	if !validIdentifier(table) {
		return Config{}, fmt.Errorf("CLICKHOUSE_TABLE has invalid value %q", table)
	}

	return Config{
		Addr:     splitCSV(getEnv("CLICKHOUSE_ADDR", "localhost:9000")),
		Database: getEnv("CLICKHOUSE_DATABASE", "default"),
		Username: getEnv("CLICKHOUSE_USERNAME", "default"),
		Password: os.Getenv("CLICKHOUSE_PASSWORD"),
		Table:    table,
	}, nil
}

func Open(cfg Config) (ch.Conn, error) {
	if len(cfg.Addr) == 0 {
		return nil, errors.New("CLICKHOUSE_ADDR cannot be empty")
	}

	conn, err := ch.Open(&ch.Options{
		Addr: cfg.Addr,
		Auth: ch.Auth{
			Database: cfg.Database,
			Username: cfg.Username,
			Password: cfg.Password,
		},
		DialTimeout: 10 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := conn.Ping(ctx); err != nil {
		return nil, err
	}

	return conn, nil
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func validIdentifier(value string) bool {
	pattern := regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*(\.[A-Za-z_][A-Za-z0-9_]*)?$`)
	return pattern.MatchString(value)
}
