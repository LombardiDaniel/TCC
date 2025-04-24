package services

type DBService interface {
	GetMacs(productId string) ([]string, error)
}

type DBServiceMock struct{}

func (db *DBServiceMock) GetMacs(productId string) ([]string, error) {
	return []string{"000000000001", "000000000002"}, nil
}
