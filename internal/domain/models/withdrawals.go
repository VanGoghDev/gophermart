package models

import "time"

type Withdrawal struct {
	ProcessedAt time.Time `json:"processed_at"`
	OrderNumber string    `json:"order"`
	Sum         int64     `json:"sum"`
}
