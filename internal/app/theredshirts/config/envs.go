package config

import "os"

var (
	//Service
	LogLevel = os.Getenv("LOG_LEVEL")

	//Database
	PostgresUser     = os.Getenv("POSTGRES_USER")
	PostgresDB       = os.Getenv("POSTGRES_DB")
	PostgresPassword = os.Getenv("POSTGRES_PASSWORD")
	PostgresHost     = os.Getenv("POSTGRES_HOST")
	PostgresPort     = os.Getenv("POSTGRES_PORT")
)
