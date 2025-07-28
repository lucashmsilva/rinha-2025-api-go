package database

import (
	"context"
	"fmt"

	"github.com/lucashmsilva/rinha-2025-api-go/internal/infra/config"
	"github.com/redis/go-redis/v9"
)

type Db struct {
	Conn *redis.Client
	Ctx  context.Context
}

func LoadConnections(dbCfg *config.DbConnCfg) *Db {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", dbCfg.Host, dbCfg.Port),
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	db := &Db{rdb, context.Background()}
	db.ResetStorage()

	return db
}

func (db *Db) ResetStorage() {
	// summary accumulators
	db.Conn.Set(db.Ctx, "summary:default:totalRequests", 0, 0)
	db.Conn.Set(db.Ctx, "summary:default:totalAmount", 0, 0)
	db.Conn.Set(db.Ctx, "summary:fallback:totalRequests", 0, 0)
	db.Conn.Set(db.Ctx, "summary:fallback:totalAmount", 0, 0)

	// healthcheck cache
	db.Conn.Set(db.Ctx, "health:default:failing", false, 0)
	db.Conn.Set(db.Ctx, "health:fallback:failing", false, 0)
}
