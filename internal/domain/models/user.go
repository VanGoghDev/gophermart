package models

type User struct {
	Login    string
	PassHash []byte
	Balance  int64
}
