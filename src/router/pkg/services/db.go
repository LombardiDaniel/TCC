package services

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type DBService interface {
	// GetRoute retrieves the current route for a given device MAC address.
	GetRoute(deviceMac string) (string, error)
}

type DBServiceRedisImpl struct {
	client *redis.Client
}

func NewDBService(client *redis.Client) DBService {

	err := client.Set(context.Background(), "arbitratyMac", "000000000000", 0).Err()
	if err != nil {
		panic(err)
	}

	return &DBServiceRedisImpl{
		client: client,
	}
}

// Here we only make a call to redis to make sure it has the appropriate delays
func (db *DBServiceRedisImpl) GetRoute(deviceMac string) (string, error) {
	// return "/gw/GW_MAC/action", nil

	res, err := db.client.Get(context.Background(), "arbitratyMac").Result()
	if err != nil {
		return "", nil
	}

	return "/gw/" + res + "/action", nil
}
