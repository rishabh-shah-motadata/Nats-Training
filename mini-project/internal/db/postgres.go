package db

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

const (
	maxConn           = 5
	minConns          = 3
	healthCheckPeriod = 3 * time.Minute
	maxConnIdleTime   = 1 * time.Minute
	maxConnLifetime   = 3 * time.Minute
	lazyConnect       = false
)

func InitPostgresDB(ctx context.Context) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	dataSourceName := "host=localhost port=5432 user=postgres password=postgres dbname=orders_db sslmode=disable"

	poolCfg, err := pgxpool.ParseConfig(dataSourceName)
	if err != nil {
		log.Println("error parsing postgres config", err)
		return nil, err
	}

	poolCfg.MaxConns = maxConn
	poolCfg.MinConns = minConns
	poolCfg.HealthCheckPeriod = healthCheckPeriod
	poolCfg.MaxConnIdleTime = maxConnIdleTime
	poolCfg.MaxConnLifetime = maxConnLifetime
	poolCfg.LazyConnect = lazyConnect

	connPool, err := pgxpool.ConnectConfig(ctx, poolCfg)
	if err != nil {
		log.Println("error connecting to postgres database", err)
		return nil, err
	}

	conn, err := connPool.Acquire(ctx)
	if err != nil {
		log.Println("error acquiring connection from postgres pool", err)
		return nil, err
	}
	defer conn.Release()

	// Ping the database
	if err = conn.Conn().Ping(ctx); err != nil {
		log.Println("error pinging postgres database", err)
		return nil, err
	}

	return connPool, nil
}
