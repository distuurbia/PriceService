// Package main contains main func and redis connection
package main

import (
	"context"
	"net"

	"github.com/caarlos0/env"
	"github.com/distuurbia/PriceService/internal/config"
	"github.com/distuurbia/PriceService/internal/handler"
	"github.com/distuurbia/PriceService/internal/repository"
	"github.com/distuurbia/PriceService/internal/service"
	protocol "github.com/distuurbia/PriceService/protocol/price"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
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
	handl := handler.NewHandler(priceServiceService)
	go priceServiceService.SendToAllSubscribedChans(context.Background())
	lis, err := net.Listen("tcp", "localhost:5433")
	if err != nil {
		logrus.Fatalf("cannot connect listener: %s", err)
	}
	serverRegistrar := grpc.NewServer()
	protocol.RegisterPriceServiceServiceServer(serverRegistrar, handl)

	err = serverRegistrar.Serve(lis)
	if err != nil {
		logrus.Fatalf("cannot serve: %s", err)
	}
}
