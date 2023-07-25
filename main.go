// Package main contains main func and redis connection
package main

import (
	"context"
	"fmt"

	"github.com/caarlos0/env"
	"github.com/distuurbia/PriceService/internal/config"
	"github.com/distuurbia/PriceService/internal/repository"
	"github.com/distuurbia/PriceService/internal/service"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

// connectRedis connects to the redis db
func connectRedis(cfg *config.Config) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddress,
		DB:   0,
	})
	return client
}

func main() {
	var cfg config.Config
	if err := env.Parse(&cfg); err != nil {
		logrus.Fatalf("main -> %v", err)
	}
	client := connectRedis(&cfg)
	priceServiceRepo := repository.NewPriceServiceRepository(client, &cfg)
	priceServiceService := service.NewPriceServiceService(priceServiceRepo)
	for {
		shares, err := priceServiceService.ReadFromStream(context.Background())
		if err != nil {
			logrus.Fatalf("main -> %v", err)
		}
		for _, share := range shares {
			fmt.Println(share)
		}
	}
}
