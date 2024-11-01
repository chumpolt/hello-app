// Copyright 2022 LINE Company (Thailand)

package main

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

func main() {
	e := echo.New()

	e.GET("/hello2", func(c echo.Context) error {
		writeSession(c.Logger())
		return c.String(http.StatusOK, "Hello-2")
	})

	e.Logger.SetLevel(log.INFO)
	e.Logger.Fatal(e.Start(":8000"))
}

func writeSession(logger echo.Logger) {
	timestamp := time.Now().UTC()
	rdb := mustRedisClient(logger)
	id := mustGetLatestSessionId(rdb, logger) + 1

	mustSetLatestSessionId(id, rdb, logger)
	logger.Infof("Session: %v - %d", timestamp, id)
}

func mustGetLatestSessionId(rdb *redis.Client, logger echo.Logger) int {
	result, err := rdb.Get(context.Background(), "latest_session_id").Result()
	if err != nil {
		if err != redis.Nil {
			logger.Fatalf("could not get latest session id, %w", err)
			os.Exit(1)
		} else {
			result = "0"
		}
	}

	id, err := strconv.Atoi(result)
	if err != nil {
		logger.Fatalf("invalid latest session, id: %v, error: %w", id, err)
		os.Exit(1)
	}

	return id
}

func mustSetLatestSessionId(id int, rdb *redis.Client, logger echo.Logger) error {
	if err := rdb.Set(context.Background(), "latest_session_id", id, 0).Err(); err != nil {
		logger.Fatalf("could not set latest session id, %w", err)
		os.Exit(1)
	}

	return nil
}

func mustRedisClient(logger echo.Logger) *redis.Client {
	db, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
		logger.Fatalf("invalid REDIS_DB, %w", err)
		os.Exit(1)
	}

	client := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       db,
	})

	return client
}
