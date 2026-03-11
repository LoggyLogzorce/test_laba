package main

import (
	"fmt"
	"log"
	"masha_laba_3/internal/db"
	"masha_laba_3/internal/handlers"
	"masha_laba_3/internal/repository"
	"masha_laba_3/internal/routers"
	"masha_laba_3/internal/services"
	"net/http"
	"time"
)

func main() {
	// Конфигурация базы данных
	// В реальном проекте эти данные стоит брать из config или .env
	dbCfg := db.Config{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "1234", // Замените на ваш пароль
		DBName:   "test_db",
		SSLMode:  "disable",
	}

	// Инициализация подключения к БД
	database, err := db.NewPostgresDB(dbCfg)
	if err != nil {
		log.Fatalf("Ошибка инициализации БД: %v", err)
	}
	defer db.CloseDB(database)

	postgresRepo := repository.NewPostgresRepo(database)
	newService := services.NewService(postgresRepo)
	h := handlers.NewHandlers(newService)

	// Настройка маршрутов
	mux := routers.SetupRoutes(h)

	// Запуск HTTP-сервера
	port := ":8080"
	fmt.Printf("Сервер запущен на http://localhost%s\n", port)
	fmt.Println("Перейдите на /register для регистрации, /login для входа.")

	srv := &http.Server{
		Addr:         port,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Не удалось запустить сервер: %v", err)
	}
}
