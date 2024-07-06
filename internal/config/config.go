package config

import "time"

type Config struct {
	Address        string        `env:"RUN_ADDRESS"`
	Env            string        `env:"ENV"`
	DSN            string        `env:"DATABASE_URI"`
	AccrualAddress string        `env:"ACCRUAL_SYSTEM_ADDRESS"`
	Secret         string        `env:"SECRET"`
	TokenExpires   time.Duration `env:"TOKEN_EXPIRES"`
}

func New() (config *Config, err error) {
	// не забыть, что в этот раз флаги имеют более высокий приоритет

	return &Config{
		Address:        "localhost:8080",
		Env:            "local",
		DSN:            "postgres://postgres:Kd100817@localhost:5440/template1?sslmode=disable",
		AccrualAddress: "",
		Secret:         "secret",
		TokenExpires:   time.Hour * 3,
	}, nil
}
