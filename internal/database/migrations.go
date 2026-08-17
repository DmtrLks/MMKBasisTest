package database

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strconv"
	"strings"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

type Migration struct {
	Version int
	Name    string
	SQL     string
}

func RunMigrations(ctx context.Context, db Database) error {
	if err := createMigrationsTable(ctx, db); err != nil {
		return fmt.Errorf("create migrations table: %w", err)
	}

	migrationsFS, err := fs.Sub(migrationFiles, "migrations")
	if err != nil {
		return fmt.Errorf("get migrations filesystem: %w", err)
	}

	migrations, err := loadMigrations(migrationsFS)
	if err != nil {
		return fmt.Errorf("load migrations: %w", err)
	}

	applied, err := getAppliedMigrations(ctx, db)
	if err != nil {
		return fmt.Errorf("get applied migrations: %w", err)
	}

	for _, migration := range migrations {
		if _, ok := applied[migration.Version]; ok {
			continue
		}

		if err := applyMigration(ctx, db, migration); err != nil {
			return fmt.Errorf("apply migration %03d_%s: %w", migration.Version, migration.Name, err)
		}
	}

	return nil
}

func createMigrationsTable(
	ctx context.Context,
	db DBTX,
) error {
	const query = `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version BIGINT UNSIGNED NOT NULL,
			applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

			PRIMARY KEY (version)
		) ENGINE=InnoDB;
	`

	_, err := db.ExecContext(ctx, query)

	return err
}

func loadMigrations(
	migrationsFS fs.FS,
) ([]Migration, error) {
	entries, err := fs.ReadDir(migrationsFS, ".")
	if err != nil {
		return nil, err
	}

	migrations := make([]Migration, 0, len(entries))

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		version, name, err := parseMigrationName(entry.Name())
		if err != nil {
			return nil, err
		}

		data, err := fs.ReadFile(migrationsFS, entry.Name())
		if err != nil {
			return nil, fmt.Errorf("read migration %s: %w", entry.Name(), err)
		}

		migrations = append(migrations, Migration{
			Version: version,
			Name:    name,
			SQL:     string(data),
		})
	}

	sort.Slice(
		migrations,
		func(i, j int) bool {
			return migrations[i].Version <
				migrations[j].Version
		},
	)

	return migrations, nil
}

func parseMigrationName(
	filename string,
) (int, string, error) {
	parts := strings.SplitN(filename, "_", 2)

	if len(parts) != 2 {
		return 0, "", fmt.Errorf("invalid migration filename %q", filename)
	}

	version, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, "", fmt.Errorf("invalid migration version in %q: %w", filename, err)
	}

	if version <= 0 {
		return 0, "", fmt.Errorf("migration version must be positive: %q", filename)
	}

	name := strings.TrimSuffix(parts[1], ".sql")

	if name == "" {
		return 0, "", fmt.Errorf("migration name is empty: %q", filename)
	}

	return version, name, nil
}

func getAppliedMigrations(
	ctx context.Context,
	db DBTX,
) (map[int]struct{}, error) {
	rows, err := db.QueryContext(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[int]struct{})

	for rows.Next() {
		var version int

		if err := rows.Scan(&version); err != nil {
			return nil, err
		}

		applied[version] = struct{}{}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return applied, nil
}

func applyMigration(
	ctx context.Context,
	db Database,
	migration Migration,
) error {
	return db.WithinTransaction(ctx, func(tx DBTX) error {
		if _, err := tx.ExecContext(ctx, migration.SQL); err != nil {
			return fmt.Errorf("execute migration: %w", err)
		}

		if _, err := tx.ExecContext(
			ctx,
			`INSERT INTO schema_migrations (version) VALUES (?)`,
			migration.Version,
		); err != nil {
			return fmt.Errorf("save migration version: %w", err)
		}

		return nil
	})
}
