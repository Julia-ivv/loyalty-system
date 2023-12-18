package main

import (
	"net/http"

	"github.com/Julia-ivv/loyalty-system.git/internal/app/accrual"
	"github.com/Julia-ivv/loyalty-system.git/internal/app/config"
	"github.com/Julia-ivv/loyalty-system.git/internal/app/handlers"
	"github.com/Julia-ivv/loyalty-system.git/internal/app/logger"
	"github.com/Julia-ivv/loyalty-system.git/internal/app/models"
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

	const numOrders = 20
	const numWorkers = 5
	ordersChan := make(chan string, numOrders)
	accrualsChan := make(chan models.ResponseAccrual, numOrders)
	accrualSystem := accrual.NewAccrualSystem(cfg.AccrualSystem, ordersChan, accrualsChan, repo)
	defer close(ordersChan)
	defer close(accrualsChan)
	for w := 1; w <= numWorkers; w++ {
		go accrualSystem.Worker()
	}
	go accrualSystem.Updater()

	err = http.ListenAndServe(cfg.Host, handlers.NewURLRouter(repo, *cfg, *accrualSystem))
	if err != nil {
		logger.ZapSugar.Fatalw(err.Error(), "event", "start server")
	}
}
