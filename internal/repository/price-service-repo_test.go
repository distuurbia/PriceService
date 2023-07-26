package repository

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/require"
)

func TestReadFromStreamEmpty(t *testing.T) {
	_, err := priceServiceRepo.ReadFromStream(context.Background())
	require.Error(t, err)
}

func TestReadFromStream(t *testing.T) {
	sharesJSON, err := json.Marshal(testShares)
	require.NoError(t, err)
	streamData := redis.XAddArgs{
		Stream: priceServiceRepo.cfg.RedisStreamName,
		Values: map[string]interface{}{
			priceServiceRepo.cfg.RedisStreamField: string(sharesJSON),
		},
	}
	_, err = priceServiceRepo.client.XAdd(context.Background(), &streamData).Result()
	require.NoError(t, err)
	shares, err := priceServiceRepo.ReadFromStream(context.Background())
	require.NoError(t, err)
	require.Equal(t, shares[0].Name, testShares[0].Name)
	require.Equal(t, shares[0].Price, testShares[0].Price)
}
