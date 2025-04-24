package services

type DBService interface {
	// GetRoute retrieves the current route for a given device MAC address.
	GetRoute(deviceMac string) (string, error)
}

type DBServiceMock struct{}

func (db *DBServiceMock) GetRoute(deviceMac string) (string, error) {
	return "GW_MAC_TOPIC", nil
}
