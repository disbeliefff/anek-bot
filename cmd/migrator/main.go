package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"os"

	"anek-bot/internal/config"

	"github.com/pressly/goose/v3"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var (
	flags = flag.NewFlagSet("goose", flag.ExitOnError)
	dir   = flags.String("dir", "migrations", "directory with migration files")
)

func main() {
	flags.Usage = usage
	flags.Parse(os.Args[1:])
	args := flags.Args()

	if len(args) < 1 {
		flags.Usage()
		os.Exit(1)
	}

	cfg, err := config.Load()
	if err != nil {
		if errors.Is(err, config.ErrEmptyBotToken) || errors.Is(err, config.ErrEmptyDBPassword) {
			fmt.Println("Note: Bot token and DB password not required for migration")
		} else {
			fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		}
	}

	dbHost := "localhost"
	dbPort := 5432
	dbUser := "anekbot"
	dbName := "anekbot"
	dbPassword := ""

	if cfg != nil {
		dbHost = cfg.Database.Host
		dbPort = cfg.Database.Port
		dbUser = cfg.Database.User
		dbName = cfg.Database.Name
		dbPassword = cfg.Database.Password
	}

	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		dbUser, dbPassword, dbHost, dbPort, dbName,
	)

	ctx := context.Background()

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to database: %v\n", err)
		os.Exit(1)
	}

	goose.SetDialect("postgres")
	goose.SetTableName("schema_migrations")

	if err := goose.RunContext(ctx, args[0], db, *dir); err != nil {
		fmt.Fprintf(os.Stderr, "Migration failed: %v\n", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Println(usagePrefix)
	flags.PrintDefaults()
	fmt.Println(usageCommands)
}

var (
	usagePrefix = `Usage: goose [OPTIONS] COMMAND

or

Set environment variables
DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME

Options:
`

	usageCommands = `
Commands:
    up                   Migrate the database to the most recent version available
    up-by-one            Migrate the database up by 1
    up-to VERSION        Migrate the database to a specific VERSION
    down                 Roll back the version by 1
    down-to VERSION      Roll back to a specific VERSION
    redo                 Re-run the latest migration
    reset                Roll back all migrations
    status               Dump the migration status
    version              Print the current version
    create NAME [sql|go] Creates new migration file with the current timestamp
    fix                  Apply sequential ordering to migrations
`
)
