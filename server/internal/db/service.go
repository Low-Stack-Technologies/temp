package db

import (
	"database/sql"
	"embed"
	_ "embed"
	"fmt"
	migrate "github.com/rubenv/sql-migrate"
	"log"
	_ "modernc.org/sqlite"
	"tech.low-stack.temp/server/internal/env"
)

//go:embed schemas/migrations/*.sql
var migrationsFs embed.FS

var databaseConnection *sql.DB

func Initialize() {
	db, err := sql.Open("sqlite", env.DatabasePath)
	if err != nil {
		panic(fmt.Errorf("failed to open database: %w", err))
	}

	migrations := &migrate.EmbedFileSystemMigrationSource{
		FileSystem: migrationsFs,
		Root:       "schemas/migrations",
	}

	n, err := migrate.Exec(db, "sqlite3", migrations, migrate.Up)
	if err != nil {
		panic(fmt.Errorf("failed to apply migrations: %w", err))
	}
	log.Printf("Applied %d migrations\n", n)

	databaseConnection = db
}

func NewQueries() *Queries {
	return New(databaseConnection)
}
