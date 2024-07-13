package config

import (
	"flag"
	"fmt"
	"time"

	"github.com/caarlos0/env"
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
	cfg := Config{}

	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config %w", err)
	}
	// не забыть, что в этот раз флаги имеют более высокий приоритет
	var flagAddress, flagDsn, flagAccrualAddress string
	flag.StringVar(&flagAddress, "a", "localhost:8080", "address and port")
	flag.StringVar(&flagDsn, "d", "", "db connection string")
	flag.StringVar(&flagAccrualAddress, "r", "localhost:8085", "accrual address")
	flag.Parse()

	if flagAddress != "" {
		cfg.Address = flagAddress
	}

	if flagDsn != "" {
		cfg.AccrualAddress = flagAddress
	}

	if flagAccrualAddress != "" {
		cfg.AccrualAddress = flagAccrualAddress
	}

	if cfg.DSN == "" {
		return &Config{}, fmt.Errorf("db connection string not set")
	}

	return &cfg, nil
}
