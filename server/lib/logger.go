package lib

import (
	"encoding/json"
	"go.uber.org/zap"
)

var logger *zap.Logger

func initLogger() *zap.Logger {
	var config zap.Config
	if GetMode() == "PRODUCTION" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
	}
	logger, _ = config.Build()
	defer logger.Sync()

	return logger
}

func PrettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}
