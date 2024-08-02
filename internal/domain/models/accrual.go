package models

type Accrual struct {
	OrderNum string      `json:"order"`
	Status   OrderStatus `json:"status"`
	Accrual  float64     `json:"accrual"`
}
