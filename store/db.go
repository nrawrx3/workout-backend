package store

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
	"github.com/nrawrx3/workout-backend/config"
	"github.com/pkg/errors"
	migrate "github.com/rubenv/sql-migrate"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func OpenSqliteDatabase(sqliteDSN string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", sqliteDSN)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open database at %s", sqliteDSN)
	}
	log.Printf("opened database at %s", sqliteDSN)
	return db, nil
}

func OpenGorm(sqliteDSN string) (*gorm.DB, error) {
	gormDB, err := gorm.Open(sqlite.Open(sqliteDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create gorm db object")
	}
	log.Printf("opened database at %s", sqliteDSN)
	return gormDB, nil
}

func RunDatabaseMigrations(cfg *config.Config) error {
	migrations := &migrate.FileMigrationSource{
		Dir: cfg.MigrationsPath,
	}

	db, err := OpenSqliteDatabase(cfg.Sqlite.SqliteDSN())
	if err != nil {
		return errors.Wrapf(err, "failed to open database at %s", cfg.Sqlite.SqliteDSN())
	}

	n, err := migrate.Exec(db, "sqlite3", migrations, migrate.Up)
	if err != nil {
		return errors.Wrapf(err, "failed to migrate")
	}

	log.Printf("applied %d migrations!", n)
	return nil
}

func RunDatabaseRollback(cfg *config.Config) error {
	migrations := &migrate.FileMigrationSource{
		Dir: cfg.MigrationsPath,
	}

	db, err := OpenSqliteDatabase(cfg.Sqlite.SqliteDSN())
	if err != nil {
		return errors.Wrapf(err, "failed to open dataase at %s", cfg.Sqlite.SqliteDSN())
	}

	n, err := migrate.Exec(db, "sqlite3", migrations, migrate.Down)
	if err != nil {
		return errors.Wrapf(err, "failed to rollback")
	}

	log.Printf("applied %d rollback(s)!", n)
	return nil
}
