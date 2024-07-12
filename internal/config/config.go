package config

import (
	"flag"
	"time"
)

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
	var flagAddress, flagDsn, flagAccrualAddress string
	flag.StringVar(&flagAddress, "a", "localhost:8080", "address and port")
	flag.StringVar(&flagDsn, "d", "", "db connection string")
	flag.StringVar(&flagAccrualAddress, "r", "localhost:8085", "accrual address")
	flag.Parse()

	cfg := Config{}

	if flagAddress != "" {
		cfg.Address = flagAddress
	}

	if flagDsn != "" {
		cfg.AccrualAddress = flagAddress
	}

	if flagAccrualAddress != "" {
		cfg.AccrualAddress = flagAccrualAddress
	}

	return &cfg, nil
}
