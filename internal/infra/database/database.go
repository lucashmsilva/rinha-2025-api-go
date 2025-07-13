package database

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lucashmsilva/rinha-2025-api-go/internal/infra/config"
)

type Db struct {
	Conn *pgxpool.Pool
}

func LoadConnections(dbCfg *config.DbConnCfg) *Db {
	// postgres://username:password@localhost:5432/database_name?option_name=value
	connectionStringFmt := "postgres://%s:%s@%s:%d/%s?pool_max_conn_lifetime=%s&pool_min_idle_conns=%d&pool_max_conns=%d&pool_min_conns=%d"

	dataSource := fmt.Sprintf(connectionStringFmt,
		dbCfg.Username,
		dbCfg.Password,
		dbCfg.Host,
		dbCfg.Port,
		dbCfg.Database,
		dbCfg.PoolMaxLifetime.String(),
		dbCfg.PoolMinIdleConns,
		dbCfg.PoolMaxOpenConns,
		dbCfg.PoolMinOpenConns,
	)

	db, err := pgxpool.New(context.Background(), dataSource)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		os.Exit(1)
	}

	return &Db{db}
}
