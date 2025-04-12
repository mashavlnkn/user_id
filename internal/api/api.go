package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"simple-service/internal/dto"

	"simple-service/internal/api/middleware"
	"simple-service/internal/service"
)

// Routers - структура для хранения зависимостей роутов
type Routers struct {
	Service service.Service
}

// NewRouters - конструктор для настройки API
func NewRouters(service service.Service, token string) *fiber.App {
	app := fiber.New()

	// Настройка CORS (разрешенные методы, заголовки, авторизация)
	app.Use(cors.New(cors.Config{
		AllowMethods:  "GET, POST, PUT, DELETE",
		AllowHeaders:  "Accept, Authorization, Content-Type, X-CSRF-Token, X-REQUEST-ID",
		ExposeHeaders: "Link",
		MaxAge:        300,
	}))

	// Группа маршрутов с авторизацией
	apiGroup := app.Group("/v1", middleware.Authorization(token))

	// Роут для создания задачи
	apiGroup.Post("/create_task", service.CreateTask)
	apiGroup.Get("/tasks/:id", service.GetTaskByID)
	apiGroup.Put("/tasks/:id", service.UpdateTask)
	apiGroup.Delete("/tasks/:id", service.DeleteTask)
	apiGroup.Get("/task_user/:id", service.GetTasksByUserID)
	app.Use(func(ctx *fiber.Ctx) error {
		// Если ошибка произошла на уровне маршрутов
		if err := ctx.Next(); err != nil {
			// Проверка ошибки
			switch err.(type) {
			case *fiber.Error:
				// Если ошибка fiber.Error, можно использовать встроенные методы для ответа с ошибкой
				return err
			default:
				// Для остальных ошибок возвращаем 500 Internal Server Error
				return dto.InternalServerError(ctx)
			}
		}
		return nil
	})

	return app
}
