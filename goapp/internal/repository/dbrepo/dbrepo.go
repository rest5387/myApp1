package dbrepo

import (
	"database/sql"

	"github.com/go-redis/cache/v9"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"github.com/rest5387/myApp1/goapp/internal/config"
	"github.com/rest5387/myApp1/goapp/internal/repository"
)

type postgresDBRepo struct {
	App *config.AppConfig
	DB  *sql.DB
}

func NewPostgresRepo(conn *sql.DB, a *config.AppConfig) repository.SQLDatabaseRepo {
	return &postgresDBRepo{
		App: a,
		DB:  conn,
	}
}

type neo4jRepo struct {
	App   *config.AppConfig
	Neo4j neo4j.Driver
}

func NewNeo4jRepo(driver neo4j.Driver, a *config.AppConfig) repository.Neo4jRepo {
	return &neo4jRepo{
		App:   a,
		Neo4j: driver,
	}
}

type redisRepo struct {
	App        *config.AppConfig
	RedisCache *cache.Cache
}

func NewRedisRepo(cache *cache.Cache, a *config.AppConfig) repository.RedisRepo {
	return &redisRepo{
		App:        a,
		RedisCache: cache,
	}
}
