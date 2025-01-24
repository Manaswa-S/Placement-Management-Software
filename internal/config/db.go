package config

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	sqlc "go.mod/internal/sqlc/generate"
)


var	Pool *pgxpool.Pool
var	QueriesPool *sqlc.Queries
var	RedisClient *redis.Client


func InitDB() (error) {
	fmt.Println("Connecting to Databases and Cache...")
	// create context object
	ctx := context.Background()
	
	// initialize database
	dbConn := os.Getenv("PMSDBLoginCredentials")
	pool, err := pgxpool.New(ctx, dbConn)
	if err != nil {
		return fmt.Errorf("error creating database pool: %s", err)
	}

	// inittialize queries pool
	QueriesPool = sqlc.New(pool)
	
	// Connect to redis client
	RedisClient = redis.NewClient(&redis.Options{
		Addr: os.Getenv("RedisAddress"),
		Password: os.Getenv("RedisPassword"),
		DB: 0,
		Protocol: 2,
	})
	_, err = RedisClient.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %v", err)
	}
	fmt.Println("Redis connection is alive!")

	conn, err := pool.Acquire(context.Background())
	if err != nil {
		return fmt.Errorf("pgx Pool connection failed: %v", err)
	}
	err = conn.Ping(context.Background())
	if err != nil {
		return fmt.Errorf("database connection failed: %v", err)
	} else {
		fmt.Println("Database connection is alive!")
	}


	return nil
}

// Close DB and Redis connections
func Close() (error) {
	fmt.Println("Closing connections to Databases and Cache...")
	if Pool != nil {
		Pool.Close()
	}
	if RedisClient != nil {
		return RedisClient.Close()
	}
	return nil
}