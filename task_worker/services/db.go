package services

type DBService interface {
	GetMacs(productId string) ([]string, error)
}

type DBServiceMock struct{}

func (db *DBServiceMock) GetMacs(productId string) ([]string, error) {
	return []string{"00:00:00:00:00:01", "00:00:00:00:00:02"}, nil
}
