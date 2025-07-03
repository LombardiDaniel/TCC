package services

type DBService interface {
	// GetRoute retrieves the current route for a given device MAC address.
	GetRoute(deviceMac string) (string, error)
}

type DBServiceMock struct{}

func (db *DBServiceMock) GetRoute(deviceMac string) (string, error) {
	// return "/gw/GW_MAC/action", nil
	return "/gw/000000000000/action", nil
}
