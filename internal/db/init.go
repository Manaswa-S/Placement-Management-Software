package db

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	sqlc "go.mod/sqlc/generate"
)


var	Pool *pgxpool.Pool
var	QueriesPool *sqlc.Queries
var	RedisClient *redis.Client


func InitDB()  {
	fmt.Println("in initdb")
	// create context object
	ctx := context.Background()
	
	// initialize database
	dbConn := os.Getenv("PMSDBLoginCredentials")
	pool, err := pgxpool.New(ctx, dbConn)
	if err != nil {
		fmt.Println("Error creating database pool: ", err)
	}

	// inittialize queries pool
	QueriesPool = sqlc.New(pool)
	
	// Connect to redis client
	RedisClient = redis.NewClient(&redis.Options{
		Addr: "localhost:5676",
		Password: "",
		DB: 0,
		Protocol: 2,
	})
}

// Close DB and Redis connections
func Close() {
	if Pool != nil {
		Pool.Close()
	}
	if RedisClient != nil {
		RedisClient.Close()
	}
}