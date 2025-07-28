package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port                 int
	LogLevel             int
	DefaultProcessorURL  string
	FallBackProcessorURL string
	ProcessorAPIToken    string
	StartHealthChecker   bool
	DbConnCfg            *DbConnCfg
}

type DbConnCfg struct {
	Host string
	Port int
}

func LoadConfig() *Config {
	appPort, _ := strconv.Atoi(os.Getenv("APP_PORT"))
	logLevel, _ := strconv.Atoi(os.Getenv("LOG_LEVEL"))

	dbPort, _ := strconv.Atoi(os.Getenv("DB_PORT"))

	dbCfg := &DbConnCfg{
		Host: os.Getenv("DB_HOST"),
		Port: dbPort,
	}

	return &Config{
		Port:                 appPort,
		LogLevel:             logLevel,
		DefaultProcessorURL:  os.Getenv("DEFAULT_PROCESSOR_URL"),
		FallBackProcessorURL: os.Getenv("FALLBACK_PROCESSOR_URL"),
		ProcessorAPIToken:    os.Getenv("PROCESSOR_API_TOKEN"),
		StartHealthChecker:   os.Getenv("START_HEALTH_CHECKER") == "1",
		DbConnCfg:            dbCfg,
	}
}
