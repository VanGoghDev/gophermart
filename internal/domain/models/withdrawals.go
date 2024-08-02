package models

import "time"

type Withdrawal struct {
	ProcessedAt         time.Time `json:"-"`
	ProcessedAtFormated string    `json:"processed_at"`
	OrderNumber         string    `json:"order"`
	Sum                 float64   `json:"sum"`
}
