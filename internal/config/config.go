package config

type Config struct {
	Address        string `env:"RUN_ADDRESS"`
	Env            string `env:"ENV"`
	DSN            string `env:"DATABASE_URI"`
	AccrualAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
}

func New() (config *Config, err error) {
	// не забыть, что в этот раз флаги имеют более высокий приоритет

	return &Config{
		Address:        "localhost:8080",
		Env:            "local",
		DSN:            "",
		AccrualAddress: "",
	}, nil
}
