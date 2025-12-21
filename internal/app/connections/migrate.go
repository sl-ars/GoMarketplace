package connections

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/golang-migrate/migrate/v4"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// convertDSNToURL converts PostgreSQL DSN from key-value format to URL format
func convertDSNToURL(dsn string) string {
	// If already in URL format, return as is
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		return dsn
	}

	// Parse key-value pairs
	params := make(map[string]string)
	re := regexp.MustCompile(`(\w+)=([^\s]+)`)
	matches := re.FindAllStringSubmatch(dsn, -1)
	for _, match := range matches {
		if len(match) == 3 {
			params[match[1]] = match[2]
		}
	}

	// Build URL
	host := params["host"]
	port := params["port"]
	user := params["user"]
	password := params["password"]
	dbname := params["dbname"]
	sslmode := params["sslmode"]

	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "5432"
	}

	url := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, password, host, port, dbname)
	if sslmode != "" {
		url += "?sslmode=" + sslmode
	}

	return url
}

func RunMigrations(dsn string) error {
	dsnURL := convertDSNToURL(dsn)
	m, err := migrate.New("file://migrations", dsnURL)
	if err != nil {
		return fmt.Errorf("create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("running migrations: %w", err)
	}

	return nil
}
