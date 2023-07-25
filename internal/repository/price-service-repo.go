// Package repository contains methods that work with Redis Stream
package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/distuurbia/PriceService/internal/config"
	"github.com/distuurbia/PriceService/internal/model"
	"github.com/go-redis/redis/v8"
)

// PriceServiceRepository contains redis client
type PriceServiceRepository struct {
	client *redis.Client
	cfg    *config.Config
}

// NewPriceServiceRepository creates and returns a new instance of PriceServiceRepository, using the provided redis.Client
func NewPriceServiceRepository(client *redis.Client, cfg *config.Config) *PriceServiceRepository {
	return &PriceServiceRepository{
		client: client,
		cfg:    cfg,
	}
}

// ReadFromStream reads last message from the redis stream
func (rdsStream *PriceServiceRepository) ReadFromStream(ctx context.Context) (shares []*model.Share, err error) {
	results, err := rdsStream.client.XRevRange(ctx, rdsStream.cfg.RedisStreamName, "+", "-").Result()
	if err != nil {
		return nil, fmt.Errorf("PriceServiceRepository -> ReadFromStream -> XRead -> %w", err)
	}
	err = json.Unmarshal([]byte(results[0].Values[rdsStream.cfg.RedisStreamField].(string)), &shares)
	if err != nil {
		return nil, fmt.Errorf("PriceServiceRepository -> ReadFromStream -> json.Unmarshal -> %w", err)
	}

	return shares, nil
}
