package lib

import (
	"go.uber.org/zap"
	"os"
)

type Config struct {
	logger *zap.Logger
	mode   string
	WP     *WorkerPoolImpl
}

var config = Config{}

func InitConfiguration() {
	envMode := os.Getenv("MODE")
	if envMode == "PRODUCTION" {
		config.mode = "PRODUCTION"
	} else if envMode == "DEVELOPMENT" {
		config.mode = "DEVELOPMENT"
	} else if config.mode == "TEST" {
		config.mode = "TEST"
	} else {
		config.mode = "DEVELOPMENT"
	}
}

func GetConfig() *Config {
	return &config
}

func GetLogger() *zap.Logger {
	if config.logger == nil {
		config.logger = initLogger()
	}
	return config.logger
}

func GetMode() string {
	return config.mode
}
