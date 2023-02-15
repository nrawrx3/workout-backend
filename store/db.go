package store

import (
	"database/sql"

	"github.com/rs/zerolog/log"

	_ "github.com/mattn/go-sqlite3"
	"github.com/nrawrx3/uno-backend/config"
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
	log.Info().Str("dsn", sqliteDSN).Msg("opened database with sql.Open")
	return db, nil
}

func OpenGorm(sqliteDSN string) (*gorm.DB, error) {
	gormDB, err := gorm.Open(sqlite.Open(sqliteDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create gorm db object")
	}
	log.Info().Str("dsn", sqliteDSN).Msg("opened database with gorm.Open")
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

	log.Info().Str("store", "migrations applied!").Int("count", n).Send()
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

	log.Info().Str("store", "rollbacks applied!").Int("count", n).Send()
	return nil
}
