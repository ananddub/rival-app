package connection

import (
	"context"
	"fmt"
	"time"

	"rival/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

var dbConnection *pgxpool.Pool

func GetPgConnection(config *config.DatabaseConfig) (*pgxpool.Pool, error) {
	if dbConnection != nil {
		return dbConnection, nil
	}
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s",
		config.User, config.Password, config.Host, config.Port, config.DBName,
	)
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	poolConfig.MaxConns = 20
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = 30 * time.Minute
	poolConfig.MaxConnIdleTime = 5 * time.Minute
	poolConfig.HealthCheckPeriod = 1 * time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, err
	}

	dbConnection = pool
	return pool, nil
}
