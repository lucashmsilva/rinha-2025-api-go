package config

import (
	"os"
	"strconv"
	"time"
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
	Database         string
	Host             string
	Port             int
	Username         string
	Password         string
	PoolMaxLifetime  time.Duration
	PoolMinIdleConns int
	PoolMaxOpenConns int
	PoolMinOpenConns int
}

func LoadConfig() *Config {
	appPort, _ := strconv.Atoi(os.Getenv("APP_PORT"))
	logLevel, _ := strconv.Atoi(os.Getenv("LOG_LEVEL"))

	dbPort, _ := strconv.Atoi(os.Getenv("DB_PORT"))
	poolMaxLifetime, _ := strconv.Atoi(os.Getenv("DB_POOL_MAX_LIFETIME"))
	poolMaxIdleConns, _ := strconv.Atoi(os.Getenv("DB_POOL_MIN_IDLE_CONNS"))
	poolMaxOpenConns, _ := strconv.Atoi(os.Getenv("DB_POOL_MAX_OPEN_CONNS"))
	poolMinOpenConns, _ := strconv.Atoi(os.Getenv("DB_POOL_MIN_OPEN_CONNS"))

	dcCfg := &DbConnCfg{
		Database:         os.Getenv("APP_DATABASE"),
		Host:             os.Getenv("DB_HOST"),
		Username:         os.Getenv("DB_USERNAME"),
		Password:         os.Getenv("DB_PASSWORD"),
		Port:             dbPort,
		PoolMaxLifetime:  time.Duration(poolMaxLifetime) * time.Second,
		PoolMinIdleConns: poolMaxIdleConns,
		PoolMaxOpenConns: poolMaxOpenConns,
		PoolMinOpenConns: poolMinOpenConns,
	}

	return &Config{
		Port:                 appPort,
		LogLevel:             logLevel,
		DefaultProcessorURL:  os.Getenv("DEFAULT_PROCESSOR_URL"),
		FallBackProcessorURL: os.Getenv("FALLBACK_PROCESSOR_URL"),
		ProcessorAPIToken:    os.Getenv("PROCESSOR_API_TOKEN"),
		StartHealthChecker:   os.Getenv("START_HEALTH_CHECKER") == "1",
		DbConnCfg:            dcCfg,
	}
}
