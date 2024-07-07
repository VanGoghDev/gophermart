package models

type Balance struct {
	Current   float64 `json:"current"`
	Withdrawn int64   `json:"withdrawn"`
}
