package services

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

const ttl = 60 * time.Second

// SharedMemoryService is an interface that defines methods for saving and retrieving
// repKey for a given deviceMac in shared memory (used only for routing).
type SharedMemoryService interface {
	// Save saves the repKey for the given deviceMac in shared memory.
	Save(ctx context.Context, deviceMac string, repKey string) error

	// RepKey retrieves the repKey for the given deviceMac from shared memory.
	RepKey(ctx context.Context, deviceMac string) (string, error)
}

type SharedMemoryServiceImpl struct {
	client *redis.Client
}

func NewSharedMemoryService(client *redis.Client) SharedMemoryService {
	return &SharedMemoryServiceImpl{
		client: client,
	}
}

func (s *SharedMemoryServiceImpl) Save(ctx context.Context, deviceMac string, repKey string) error {
	return s.client.SetEx(ctx, deviceMac, repKey, ttl).Err()
}

func (s *SharedMemoryServiceImpl) RepKey(ctx context.Context, deviceMac string) (string, error) {
	panic("not implemented")
}
