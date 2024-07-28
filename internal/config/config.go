package config

import (
	"errors"
	"flag"
	"fmt"
	"time"

	"github.com/caarlos0/env"
)

type Config struct {
	Address             string        `env:"RUN_ADDRESS"`
	Env                 string        `env:"ENV"`
	DSN                 string        `env:"DATABASE_URI"`
	AccrualAddress      string        `env:"ACCRUAL_SYSTEM_ADDRESS"`
	Secret              string        `env:"SECRET"`
	TokenExpires        time.Duration `env:"TOKEN_EXPIRES"`
	AccrualTimeout      time.Duration `env:"ACCRUALL_TIMEOUT"`
	AccrualRetryTimeout time.Duration `env:"ACCRUAL_RETRY_TIMEOUT"`
	WorkersCount        int32         `env:"WORKERS_COUNT"`
}

func New() (config *Config, err error) {
	cfg := Config{}

	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config %w", err)
	}

	var flagAddress, flagDsn, flagAccrualAddress, flagSecret string
	var flagTokenExpires, defaultTokenLifeTime, flagAccrualTimeout, defaultAccrualTimeout,
		flagWorkersCount, flagAccrualRetryTimeout int64
	defaultTokenLifeTime = 3
	defaultAccrualTimeout = 3
	flag.StringVar(&flagAddress, "a", "", "address and port")
	flag.StringVar(&flagDsn, "d", "", "db connection string")
	flag.StringVar(&flagAccrualAddress, "r", "", "accrual address")
	flag.StringVar(&flagSecret, "s", "secret", "token secret")
	flag.Int64Var(&flagTokenExpires, "e", defaultTokenLifeTime, "token expires (hours)")
	flag.Int64Var(&flagAccrualTimeout, "t", defaultAccrualTimeout, "timeout for accrual requests (seconds)")
	flag.Int64Var(&flagWorkersCount, "w", 1, "number of workers")
	flag.Int64Var(&flagAccrualRetryTimeout, "t", defaultAccrualTimeout,
		"timeout for retry accrual requests  after 428(seconds)")

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
		cfg.AccrualTimeout = time.Second * time.Duration(flagAccrualTimeout)
	}

	if flagTokenExpires > 0 {
		cfg.TokenExpires = time.Hour * time.Duration(flagTokenExpires)
	}

	if flagWorkersCount > 0 {
		cfg.WorkersCount = int32(flagWorkersCount)
	}

	if flagAccrualRetryTimeout > 0 {
		cfg.AccrualRetryTimeout = time.Second * time.Duration(flagAccrualRetryTimeout)
	}

	if cfg.DSN == "" {
		return &Config{}, errors.New("db connection string not set")
	}

	return &cfg, nil
}
