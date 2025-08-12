package middleware

import "github.com/redis/go-redis/v9"

var redisClient *redis.Client

func SetRedisClient(client *redis.Client) {
    redisClient = client
}
