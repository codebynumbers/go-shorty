package configuration

type Config struct {
	HostPort  string `envconfig:"HOST_PORT" default:"3000"`
	Domain    string `envconfig:"DOMAIN" default:"localhost"`
	DbDriver  string `envconfig:"DB_DRIVER" default:"sqlite3"`
	DbPath    string `envconfig:"DB_PATH" default:"./shorty.db"`
	RedisHost string `envconfig:"REDIS_HOST" default:"localhost"`
	RedisPort string `envconfig:"REDIS_PORT" default:"16379"`
}
