package db

import (
	"context"
	"embed"
	"errors"
	"fmt"

	"github.com/BeanCodeDe/TheRedShirts-Lobby/internal/app/theredshirts/util"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	_ "github.com/lib/pq"
)

var (
	//go:embed migration/postgres/*.up.sql
	postgresMigrationFs embed.FS
)

type (
	postgresConnection struct {
		dbPool *pgxpool.Pool
	}

	postgresTransaction struct {
		tx pgx.Tx
	}
)

func newPostgresConnection() (DB, error) {
	user := util.GetEnvWithFallback("POSTGRES_USER", "postgres")
	dbName := util.GetEnvWithFallback("POSTGRES_DB", "postgres")
	password, err := util.GetEnv("POSTGRES_PASSWORD")
	if err != nil {
		return nil, fmt.Errorf("postgres password has to be set: %v", err)
	}
	host := util.GetEnvWithFallback("POSTGRES_HOST", "postgres")
	port, err := util.GetEnvIntWithFallback("POSTGRES_PORT", 5432)
	options := util.GetEnvWithFallback("POSTGRES_OPTIONS", "sslmode=disable")
	migrationOptions := util.GetEnvWithFallback("POSTGRES_MIGRATION_OPTIONS", "&x-migrations-table=theredshirts-lobby")

	if err != nil {
		return nil, fmt.Errorf("port is not a number: %v", err)
	}

	url := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?%s", user, password, host, port, dbName, options)
	err = migratePostgresDatabase(url + migrationOptions)
	if err != nil {
		return nil, fmt.Errorf("error while migrating database: %v", err)
	}

	dbPool, err := pgxpool.Connect(context.Background(), url)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %v", err)
	}
	return &postgresConnection{dbPool: dbPool}, nil
}

func (connection *postgresConnection) Close() {
	connection.dbPool.Close()
}

func migratePostgresDatabase(url string) error {
	d, err := iofs.New(postgresMigrationFs, "migration/postgres")
	if err != nil {
		return fmt.Errorf("error while creating instance of migration scrips: %v", err)
	}
	m, err := migrate.NewWithSourceInstance("iofs", d, url)
	if err != nil {
		return fmt.Errorf("error while creating instance of migration scrips: %v", err)
	}
	err = m.Up()
	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			return nil
		}
		return fmt.Errorf("error while migrating: %v", err)
	}
	return nil
}

func (db *postgresConnection) StartTransaction() (DBTx, error) {
	tx, err := db.dbPool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("unknown error while starting transaction: %v", err)
	}
	return &postgresTransaction{tx: tx}, nil

}

func (tx *postgresTransaction) Commit() error {
	return tx.tx.Commit(context.Background())
}

func (tx *postgresTransaction) Rollback() error {
	return tx.tx.Rollback(context.Background())
}
