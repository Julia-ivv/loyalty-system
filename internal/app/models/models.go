package models

import (
	"time"
)

type OrderStatus string

const (
	NewOrder        OrderStatus = "NEW"
	OrderRegistered OrderStatus = "REGISTERED"
	OrderProcessing OrderStatus = "PROCESSING"
	OrderInvalid    OrderStatus = "INVALID"
	OrderProcessed  OrderStatus = "PROCESSED"
)

type RequestRegData struct {
	Login string `json:"login"`
	Pwd   string `json:"password"`
}

type RequestAuthData struct {
	Login string `json:"login"`
	Pwd   string `json:"password"`
}

type ResponseOrder struct {
	Number       string    `json:"number"`
	Status       string    `json:"status"`
	Accrual      float32   `json:"accrual,omitempty"`
	UploadedTime time.Time `json:"uploaded_at"`
}

type ResponseBalance struct {
	PointsBalance float32 `json:"current"`
	PointsUsed    float32 `json:"withdrawn"`
}

type RequestWithdrawData struct {
	OrderNumber string  `json:"order"`
	WithdrawSum float32 `json:"sum"`
}

type ResponseWithdrawals struct {
	OrderNumber  string    `json:"order"`
	WithdrawSum  float32   `json:"sum"`
	WithdrawTime time.Time `json:"processed_at"`
}

type ResponseAccrual struct {
	OrderNumber string      `json:"order"`
	OrderStatus OrderStatus `json:"status"`
	Accrual     float32     `json:"accrual"`
}
