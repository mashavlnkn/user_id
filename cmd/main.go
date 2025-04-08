package main

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"log"
	"os"
	"os/signal"
	"simple-service/internal/api"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"

	"simple-service/internal/config"
	customLogger "simple-service/internal/logger"
	"simple-service/internal/repo"
	"simple-service/internal/service"
)

func main() {
	if err := godotenv.Load(config.EnvPath); err != nil {
		log.Fatal("Ошибка загрузки env файла:", err)
	}

	// Загружаем конфигурацию из переменных окружения
	var cfg config.AppConfig
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatal(errors.Wrap(err, "failed to load configuration"))
	}

	// Инициализация логгера
	logger, err := customLogger.NewLogger(cfg.LogLevel)
	if err != nil {

		log.Fatal(errors.Wrap(err, "error initializing logger"))
	}

	// Подключение к PostgreSQL
	repository, err := repo.NewRepository(context.Background(), cfg.PostgreSQL)
	if err != nil {
		logger.Fatal(errors.Wrap(err, "failed to initialize repository"))
	}

	// Создание сервиса с бизнес-логикой
	serviceInstance := service.NewService(repository, logger)

	// Инициализация API
	app := api.NewRouters(serviceInstance, cfg.AuthToken)
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Сервер работает!")
	})

	app.Get("/tasks/:id", serviceInstance.GetTaskByID)
	app.Post("/create_task", serviceInstance.CreateTask)
	app.Get("/tasks", serviceInstance.GetAllTasks)    // Получить список всех задач
	app.Put("/tasks/:id", serviceInstance.UpdateTask) // Обновить задачу
	app.Delete("/tasks/:id", serviceInstance.DeleteTask)

	// Запуск HTTP-сервера в отдельной горутине
	go func() {
		logger.Infof("Starting server on %s", cfg.Rest.ListenAddress)
		if err := app.Listen(cfg.Rest.ListenAddress); err != nil {
			log.Fatal(errors.Wrap(err, "failed to start server"))
		}
	}()

	// Ожидание системных сигналов для корректного завершения работы
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	<-signalChan

	logger.Info("Shutting down gracefully...")
	if err := app.Shutdown(); err != nil {
		logger.Fatal("Error during shutdown", zap.Error(err))
	}

	logger.Info("Server stopped")
}
