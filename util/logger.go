package util

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var Logger *zap.Logger

func InitLogger() {
  if viper.GetString("environment") == "production" {
    Logger, _ = zap.NewProduction()
  } else {
    Logger, _ = zap.NewDevelopment()
  }
}
