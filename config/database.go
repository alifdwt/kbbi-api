package config

import (
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

func NewDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5436"),
		User:     getEnv("DB_USER", "kbbi_user"),
		Password: getEnv("DB_PASSWORD", "kbbi_password"),
		DBName:   getEnv("DB_NAME", "kbbi"),
	}
}

func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.Host, c.Port, c.User, c.Password, c.DBName)
}

func ConnectDB(config *DatabaseConfig) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", config.GetDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return db, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func GetEnv(key, defaultValue string) string {
	return getEnv(key, defaultValue)
}