package config

import "os"

type DB struct {
	Host     string
	User     string
	Password string
	DBName   string
	Port     string
}

type Server struct {
	Port string
}

type Config struct {
	DB           DB
	Server       Server
	JWTSecretKey string
}

func GetConfig() Config {
	return Config{
		DB: DB{
			Host:     os.Getenv("DB_HOST"),
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
			Port:     os.Getenv("DB_PORT"),
			DBName:   os.Getenv("DB_NAME"),
		},
		Server: Server{
			Port: os.Getenv("SERVER_PORT"),
		},
		JWTSecretKey: os.Getenv("JWT_SECRET_KEY"),
	}
}
