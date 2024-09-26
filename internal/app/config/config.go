package config

import (
	"flag"

	"github.com/caarlos0/env"
)

type Flags struct {
	Host          string `env:"RUN_ADDRESS"`            // -a адрес запуска сервиса, например localhost:8080
	DBURI         string `env:"DATABASE_URI"`           // -d строка с адресом подключения к БД
	AccrualSystem string `env:"ACCRUAL_SYSTEM_ADDRESS"` // -r адрес системы расчета начислений, http://localhost:9090
}

const (
	NumOrders  = 20
	NumWorkers = 5
)

func NewConfig() *Flags {
	c := &Flags{}

	flag.StringVar(&c.Host, "a", ":8080", "HTTP server start address")
	flag.StringVar(&c.DBURI, "d", "", "database connection address")
	flag.StringVar(&c.AccrualSystem, "r", "", "accrual system address")
	flag.Parse()

	env.Parse(c)

	return c
}
