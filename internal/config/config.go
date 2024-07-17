package config

import (
	"errors"
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
	AccrualTimeout time.Duration `env:"ACCRUALL_TIMEOUT"`
}

func New() (config *Config, err error) {
	cfg := Config{}

	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config %w", err)
	}

	var flagAddress, flagDsn, flagAccrualAddress, flagSecret string
	var flagTokenExpires, defaultTokenLifeTime, flagAccrualTimeout, defaultAccrualTimeout int64
	defaultTokenLifeTime = 3
	defaultAccrualTimeout = 3
	flag.StringVar(&flagAddress, "a", "", "address and port")
	flag.StringVar(&flagDsn, "d", "", "db connection string")
	flag.StringVar(&flagAccrualAddress, "r", "", "accrual address")
	flag.StringVar(&flagSecret, "s", "secret", "token secret")
	flag.Int64Var(&flagTokenExpires, "t", defaultTokenLifeTime, "token expires (hours)")
	flag.Int64Var(&flagAccrualTimeout, "t", defaultAccrualTimeout, "timeout for accrual requests (minutes)")
	flag.Parse()

	if flagAddress != "" {
		cfg.Address = flagAddress
	}

	if flagDsn != "" {
		cfg.DSN = flagDsn
	}

	if flagAccrualAddress != "" {
		cfg.AccrualAddress = flagAccrualAddress
	}

	if flagSecret != "" {
		cfg.Secret = flagSecret
	}

	if flagAccrualTimeout > 0 {
		cfg.AccrualTimeout = time.Minute * time.Duration(flagAccrualTimeout)
	}

	if flagTokenExpires > 0 {
		cfg.TokenExpires = time.Hour * time.Duration(flagTokenExpires)
	}

	if cfg.DSN == "" {
		return &Config{}, errors.New("db connection string not set")
	}

	return &cfg, nil
}
