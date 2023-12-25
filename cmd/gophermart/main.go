package main

import (
	"net/http"

	"github.com/Julia-ivv/loyalty-system.git/internal/app/accrual"
	"github.com/Julia-ivv/loyalty-system.git/internal/app/config"
	"github.com/Julia-ivv/loyalty-system.git/internal/app/handlers"
	"github.com/Julia-ivv/loyalty-system.git/internal/app/logger"
	"github.com/Julia-ivv/loyalty-system.git/internal/app/storage"
)

func main() {
	cfg := config.NewConfig()

	logger.ZapSugar = logger.NewLogger()
	logger.ZapSugar.Infow("Starting server", "addr", cfg.Host)
	logger.ZapSugar.Infow("flags", "db dsn", cfg.DBURI)

	repo, err := storage.NewStorage(*cfg)
	if err != nil {
		logger.ZapSugar.Fatal(err)
	}
	defer repo.Close()

	ordersChan := make(chan string, config.NumOrders)
	accrualsChan := make(chan storage.ResponseAccrual, config.NumOrders)
	accrualSystem := accrual.NewAccrualSystem(cfg.AccrualSystem, ordersChan, accrualsChan, repo)
	defer close(ordersChan)
	defer close(accrualsChan)
	for w := 1; w <= config.NumWorkers; w++ {
		go accrualSystem.Worker()
	}
	go accrualSystem.Updater()

	err = http.ListenAndServe(cfg.Host, handlers.NewURLRouter(repo, *cfg, *accrualSystem))
	if err != nil {
		logger.ZapSugar.Fatalw(err.Error(), "event", "start server")
	}
}
