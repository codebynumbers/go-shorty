package configuration

import (
	"github.com/kelseyhightower/envconfig"
	"log"
)

type Config struct {
	HostPort       string `envconfig:"HOST_PORT" default:"3000"`
	HostDomain     string `envconfig:"DOMAIN" default:"localhost"`
	ExternalDomain string `envconfig:"EXTERNAL_DOMAIN" default:"localhost:3000"`
	DbDriver       string `envconfig:"DB_DRIVER" default:"sqlite3"`
	DbPath         string `envconfig:"DB_PATH" default:"./shorty.db"`
	RedisHost      string `envconfig:"REDIS_HOST" default:"localhost"`
	RedisPort      string `envconfig:"REDIS_PORT" default:"16379"`
}

func Configure() Config {
	var appConfig Config
	var err error
	if err = envconfig.Process("GOSHORTY", &appConfig); err != nil {
		log.Fatal(err)
	}
	return appConfig
}
