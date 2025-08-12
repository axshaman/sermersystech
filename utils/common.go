package utils

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	_ "github.com/go-sql-driver/mysql"
)

func GenerateRandomPassword(n int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func NullIfEmpty(value string) interface{} {
	if value == "" {
		return nil
	}
	return value
}

func GetObjectTypeID(db *sql.DB, name string) (int64, error) {
	var id int64
	err := db.QueryRow("SELECT id FROM objects_types WHERE name = ?", strings.ToLower(name)).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("object_type '%s' not found", name)
	}
	return id, nil
}

func InitRedis() *redis.Client {
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		host = "redis:6379"
	}

	dbStr := os.Getenv("REDIS_DB")
	db := 0
	if dbStr != "" {
		var err error
		db, err = strconv.Atoi(dbStr)
		if err != nil {
			log.Printf("⚠️ Invalid REDIS_DB value '%s', using default 0", dbStr)
			db = 0
		}
	}

	password := os.Getenv("REDIS_PASSWORD")

	rdb := redis.NewClient(&redis.Options{
		Addr:     host,
		Password: password,
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := rdb.Ping(ctx).Result(); err != nil {
		log.Fatalf("❌ Failed to connect to Redis at %s: %v", host, err)
	}

	log.Printf("✅ Connected to Redis at %s (DB: %d)", host, db)
	return rdb
}

func InitKafka() *kafka.Writer {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "kafka:9092"
	}

	writer := &kafka.Writer{
		Addr:     kafka.TCP(brokers),
		Topic:    "api_gateway.events",
		Balancer: &kafka.LeastBytes{},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := kafka.DialContext(ctx, "tcp", brokers)
	if err != nil {
		log.Fatalf("❌ Failed to connect to Kafka at %s: %v", brokers, err)
	}
	defer conn.Close()

	log.Printf("✅ Connected to Kafka at %s", brokers)
	return writer
}