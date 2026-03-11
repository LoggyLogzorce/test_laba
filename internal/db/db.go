package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

// Config содержит параметры подключения к базе данных
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// NewPostgresDB создает и возвращает новое подключение к базе данных PostgreSQL
func NewPostgresDB(cfg Config) (*sql.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия подключения к БД: %w", err)
	}

	// Проверяем соединение с базой данных
	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("ошибка подключения к БД: %w", err)
	}

	log.Println("Успешное подключение к базе данных PostgreSQL!")
	return db, nil
}

// CloseDB закрывает подключение к базе данных
func CloseDB(db *sql.DB) {
	if db != nil {
		err := db.Close()
		if err != nil {
			log.Printf("Ошибка при закрытии подключения к БД: %v", err)
		}
		log.Println("Подключение к БД закрыто.")
	}
}
