package models

import "time"

type OrderStatus string

const (
	New        OrderStatus = "NEW"
	Registered OrderStatus = "REGISTERED"
	Invalid    OrderStatus = "INVALID"
	Processing OrderStatus = "PROCESSING"
	Processed  OrderStatus = "PROCESSED"
)

type Order struct {
	UploadedAt         time.Time   `json:"-"`
	UploadedAtFormated string      `json:"uploaded_at"`
	Number             string      `json:"number"`
	UserLogin          string      `json:"-"`
	Status             OrderStatus `json:"status"`
	Accrual            float64     `json:"accrual,omitempty"`
}
