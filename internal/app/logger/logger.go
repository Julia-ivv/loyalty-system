package logger

import (
	"go.uber.org/zap"
)

var ZapSugar *zap.SugaredLogger

func NewLogger() *zap.SugaredLogger {
	log, errLog := zap.NewDevelopment()
	if errLog != nil {
		panic(errLog)
	}
	defer log.Sync()

	zapSugar := log.Sugar()

	return zapSugar
}
