package history

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
)

func RunMySQLMigrations(db *sql.DB, migrationPath string) error {
	data, err := os.ReadFile(migrationPath)
	if err != nil {
		return fmt.Errorf("read migration file: %w", err)
	}
	stmt := strings.TrimSpace(string(data))
	if stmt == "" {
		return nil
	}
	if _, err := db.Exec(stmt); err != nil {
		return fmt.Errorf("execute migration: %w", err)
	}
	return nil
}
