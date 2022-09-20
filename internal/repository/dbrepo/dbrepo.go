package dbrepo

import (
	"database/sql"

	"github.com/neo4j/neo4j-go-driver/neo4j"
	"github.com/rest5387/myApp1/internal/config"
	"github.com/rest5387/myApp1/internal/repository"
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
