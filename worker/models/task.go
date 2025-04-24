package models

import (
	"time"
)

type Task struct {
	Action        string    `json:"action"`
	TransactionId string    `json:"transaction_id"`
	ProductId     string    `json:"product_id"`
	Ts            time.Time `json:"ts"`
}
