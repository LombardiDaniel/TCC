package services

import (
	"github.com/redis/go-redis/v9"
)

// SharedMemoryService is an interface that defines methods for saving and retrieving
// repKey for a given deviceMac in shared memory (used only for routing).
type SharedMemoryService interface {
	// Save saves the repKey for the given deviceMac in shared memory.
	Save(deviceMac string, repKey string) error

	// RepKey retrieves the repKey for the given deviceMac from shared memory.
	RepKey(deviceMac string) (string, error)
}

type SharedMemoryServiceImpl struct {
	client *redis.Client
}

func NewSharedMemoryService(client *redis.Client) SharedMemoryService {
	return &SharedMemoryServiceImpl{
		client: client,
	}
}

func (s *SharedMemoryServiceImpl) Save(deviceMac string, repKey string) error {
	panic("not implemented")
}

func (s *SharedMemoryServiceImpl) RepKey(deviceMac string) (string, error) {
	panic("not implemented")
}
