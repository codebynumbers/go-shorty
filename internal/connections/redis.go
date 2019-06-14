package connections

import (
	"fmt"
	"github.com/codebynumbers/go-shorty/internal/configuration"
	"github.com/go-redis/redis"
)

var Cache *redis.Client

func InitRedis(config configuration.Config) {
	Cache = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", config.RedisHost, config.RedisPort),
	})
}
