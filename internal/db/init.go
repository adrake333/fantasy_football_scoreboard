package db




import (
	"database/sql"
	_ "embed"
	"fmt"

	_ "modernc.org/sqlite"
)




//go:embed schema.sql
var schemaSQL string

const pragmas = `
PRAGMA journal_mode = WAL;
PRAGMA foreign_keys = ON;
`

func Init(dbPath string) (*sql.DB, error) {
	dbConn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	if err := dbConn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping sqlite database: %w", err)
	}

	if _, err := dbConn.Exec(pragmas); err != nil {
		return nil, fmt.Errorf("failed to apply sqlite pragmas: %w", err)
	}

	if _, err := dbConn.Exec(schemaSQL); err != nil {
		return nil, fmt.Errorf("failed to execute schema: %w", err)
	}

	return dbConn, nil
}