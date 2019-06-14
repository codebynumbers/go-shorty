package connections

import (
	"fmt"
	"github.com/codebynumbers/go-shorty/internal/configuration"
	"github.com/go-redis/redis"
)

func InitRedis(config configuration.Config) *redis.Client {
	cache := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", config.RedisHost, config.RedisPort),
	})
	return cache
}
