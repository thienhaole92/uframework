//nolint:revive,exhaustruct
package postgres

import (
	"database/sql"
	"errors"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

const (
	postgresDriverName = "postgres"
)

func MigrateUp(dbURI, source string) error {
	db, err := sql.Open(postgresDriverName, dbURI)
	if err != nil {
		panic(err)
	}

	driver, err := pgx.WithInstance(db, &pgx.Config{})
	if err != nil {
		panic(err)
	}

	migrator, err := migrate.NewWithDatabaseInstance(source, postgresDriverName, driver)
	if err != nil {
		panic(err)
	}

	err = migrator.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		panic(err)
	}

	sourceErr, databaseErr := migrator.Close()

	if sourceErr != nil {
		return sourceErr
	}

	if databaseErr != nil {
		return databaseErr
	}

	return nil
}
