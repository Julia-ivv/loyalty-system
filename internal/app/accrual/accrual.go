package accrual

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/Julia-ivv/loyalty-system.git/internal/app/logger"
	"github.com/Julia-ivv/loyalty-system.git/internal/app/models"
	"github.com/Julia-ivv/loyalty-system.git/internal/app/storage"
)

type AccrualSystem struct {
	AccrualAddress string
	AccrualClient  http.Client
	OrdersChan     chan string
	AccrualsChan   chan models.ResponseAccrual
	Repo           storage.PointsWorker
}

func NewAccrualSystem(accrualAddress string, ordersChan chan string,
	accrualsChan chan models.ResponseAccrual,
	repo storage.PointsWorker) *AccrualSystem {
	return &AccrualSystem{
		AccrualAddress: accrualAddress,
		AccrualClient:  http.Client{Timeout: 6 * time.Second},
		OrdersChan:     ordersChan,
		AccrualsChan:   accrualsChan,
		Repo:           repo,
	}
}

func (as *AccrualSystem) AddOrderForWork(orderNumber string) {
	as.OrdersChan <- orderNumber
}

func (as *AccrualSystem) AddAccrualForUpdate(accrual models.ResponseAccrual) {
	as.AccrualsChan <- accrual
}

func (as *AccrualSystem) GetAccrualData(orderNumber string) (models.ResponseAccrual, error) {
	resp, err := as.AccrualClient.Get(as.AccrualAddress + "/api/orders/" + orderNumber)
	if err != nil {
		return models.ResponseAccrual{}, err
	}

	switch resp.StatusCode {
	case 200:
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return models.ResponseAccrual{}, err
		}
		var respAccrual models.ResponseAccrual
		err = json.Unmarshal(body, &respAccrual)
		if err != nil {
			return models.ResponseAccrual{}, err
		}
		return respAccrual, nil
	case 204:
		return models.ResponseAccrual{}, NewAccrualError(NotRegistered, err)
	case 429:
		return models.ResponseAccrual{}, NewAccrualError(TooManyRequests, err)
	case 500:
		return models.ResponseAccrual{}, NewAccrualError(InternalError, err)
	default:
		return models.ResponseAccrual{}, nil
	}
}

const (
	statusWaitSec        = 5
	accrualSystemWaitSec = 60
	retryAttempts        = 3
)

func (as *AccrualSystem) Worker() {
	for ordNum := range as.OrdersChan {
		go func(ordNum string) {
			i := retryAttempts
			for {
				res, err := as.GetAccrualData(ordNum)
				if err != nil {
					var accErr *AccrualErr
					if errors.As(err, &accErr) && accErr.ErrType == TooManyRequests {
						time.Sleep(accrualSystemWaitSec * time.Second)
						if i--; i > 0 {
							continue
						}
						logger.ZapSugar.Infoln("request for accrual system:", err)
						return
					}
					if errors.As(err, &accErr) && accErr.ErrType == NotRegistered {
						logger.ZapSugar.Infoln("request for accrual system:", err)
						return
					}
					if errors.As(err, &accErr) && accErr.ErrType == InternalError {
						logger.ZapSugar.Infoln("request for accrual system:", err)
						return
					}
				}
				if res.OrderNumber == "" {
					return
				}
				if (res.OrderStatus == models.OrderInvalid) || (res.OrderStatus == models.OrderProcessed) {
					as.AddAccrualForUpdate(res)
					return
				}
				if (res.OrderStatus == models.OrderRegistered) || (res.OrderStatus == models.OrderProcessing) {
					time.Sleep(statusWaitSec * time.Second)
					i = retryAttempts
					continue
				}
			}
		}(ordNum)
	}
}

func (as *AccrualSystem) Updater() {
	for accr := range as.AccrualsChan {
		go func(ra models.ResponseAccrual) {
			err := as.Repo.UpdateUserAccrual(context.Background(), ra)
			if err != nil {
				logger.ZapSugar.Infoln("update order data in storage:", err)
			}
		}(accr)
	}
}
