// Package config represents struct Config.
package config

// Config is a structure of environment variables.
type Config struct {
	RedisAddress     string `env:"REDIS_ADDRESS"`
	RedisStreamName  string `env:"REDIS_STREAM_NAME"`
	RedisStreamField string `env:"REDIS_STREAM_FIELD"`
}
