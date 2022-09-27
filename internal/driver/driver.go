package driver

import (
	"context"
	"database/sql"
	"time"

	"github.com/go-redis/cache/v9"
	"github.com/go-redis/redis/v9"
	_ "github.com/lib/pq"

	"github.com/neo4j/neo4j-go-driver/neo4j"
)

// DB holds the database connection pool
type DB struct {
	SQL        *sql.DB
	Neo4j      neo4j.Driver
	RedisCache *cache.Cache
}

var dbConn = &DB{}

const maxOpenDbConn = 10
const maxIdleDbConn = 5
const maxDbLifetime = 5 * time.Minute

// ConnectSQL creates database pool for Postgres
func ConnectSQL(dsn string) (*DB, error) {
	d, err := NewDatabase(dsn)
	if err != nil {
		panic(err)
	}

	d.SetMaxOpenConns(maxOpenDbConn)
	d.SetMaxIdleConns(maxIdleDbConn)
	d.SetConnMaxLifetime(maxDbLifetime)

	dbConn.SQL = d

	err = testDB(d)
	if err != nil {
		return nil, err
	}
	return dbConn, nil
}

// testDB tries to ping the database
func testDB(d *sql.DB) error {
	err := d.Ping()
	if err != nil {
		return err
	}
	return nil
}

// NewDatabase creates a new database for the application
func NewDatabase(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func ConnectNeo4j(dsn string, username string, password string) (*DB, error) {
	driver, err := neo4j.NewDriver(dsn, neo4j.BasicAuth(username, password, ""), func(config *neo4j.Config) {
		// set Neo4j driver/connections configs
		config.Encrypted = false
	})
	if err != nil {
		return nil, err
	}
	dbConn.Neo4j = driver
	return dbConn, nil
}

// ConnetRedis connect to Redis and get an Ring,
// then set the Ring to be a cache with TinyLFU algo.
func ConnectRedis(addr string, password string, db int) (*DB, error) {
	// context for ring.Ping()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	ring := redis.NewRing(&redis.RingOptions{
		Addrs: map[string]string{
			"server1": addr,
		},
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	_, err := ring.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	myCache := cache.New(&cache.Options{
		Redis:      ring,
		LocalCache: cache.NewTinyLFU(1000, time.Minute),
	})

	dbConn.RedisCache = myCache

	return dbConn, nil
}
