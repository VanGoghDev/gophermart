package models

type Accrual struct {
	OrderNum string      `json:"order"`
	Status   OrderStatus `json:"status"`
	Accrual  int64       `json:"accrual"`
}
