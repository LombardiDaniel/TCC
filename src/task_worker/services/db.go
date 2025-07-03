package services

import (
	"fmt"
	"math/rand"
)

type DBService interface {
	GetMacs(productId string) ([]string, error)
}

type DBServiceMock struct{}

func randomMAC() string {
	buf := make([]byte, 6)
	for i := 0; i < 6; i++ {
		buf[i] = byte(rand.Intn(256))
	}
	return fmt.Sprintf("%02x%02x%02x%02x%02x%02x", buf[0], buf[1], buf[2], buf[3], buf[4], buf[5])
}

func (db *DBServiceMock) GetMacs(productId string) ([]string, error) {
	return []string{randomMAC(), randomMAC()}, nil
}
