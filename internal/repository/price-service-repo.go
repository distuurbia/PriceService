// Package repository contains methods that work with Redis Stream
package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/distuurbia/PriceService/internal/config"
	"github.com/distuurbia/PriceService/internal/model"
	"github.com/distuurbia/PriceService/proto_services"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
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
func (priceServiceRepo *PriceServiceRepository) ReadFromStream(ctx context.Context) (shares []*model.Share, err error) {
	results, err := priceServiceRepo.client.XRevRange(ctx, priceServiceRepo.cfg.RedisStreamName, "+", "-").Result()
	if err != nil {
		return nil, fmt.Errorf("PriceServiceRepository -> ReadFromStream -> XRead -> %w", err)
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("PriceServiceRepository -> ReadFromStream -> error: message is empty")
	}

	err = json.Unmarshal([]byte(results[0].Values[priceServiceRepo.cfg.RedisStreamField].(string)), &shares)
	if err != nil {
		return nil, fmt.Errorf("PriceServiceRepository -> ReadFromStream -> json.Unmarshal -> %w", err)
	}

	return shares, nil
}

// SendToSubscriber sends the messages from redis stream to exact subscriber
func (priceServiceRepo *PriceServiceRepository) SendToSubscriber(ctx context.Context, subscriberID uuid.UUID,
	subscribersShares map[uuid.UUID]chan []*model.Share, stream proto_services.PriceServiceService_SubscribeServer) {
	for {
		select {
		case <-ctx.Done():
			return
		case shares := <-subscribersShares[subscriberID]:
			var protoShares []*proto_services.Share
			for _, share := range shares {
				protoShares = append(protoShares, &proto_services.Share{
					Name:  share.Name,
					Price: share.Price,
				})
			}

			err := stream.Send(&proto_services.SubscribeResponse{Shares: protoShares})
			if err != nil {
				logrus.Errorf("PriceServiceRepository -> SendToSubscriber -> stream.Send: %v", err)
				return
			}
		}
	}
}
