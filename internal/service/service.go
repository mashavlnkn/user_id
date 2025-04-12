package service

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"simple-service/internal/dto"
	"simple-service/internal/repo"
	"simple-service/pkg/validator"
	"strconv"
)

// Слой бизнес-логики. Тут должна быть основная логика сервиса

// Service - интерфейс для бизнес-логики
type Service interface {
	CreateTask(ctx *fiber.Ctx) error
	GetTaskByID(ctx *fiber.Ctx) error
	UpdateTask(ctx *fiber.Ctx) error
	DeleteTask(ctx *fiber.Ctx) error
	GetTasksByUserID(ctx *fiber.Ctx) error
}

type service struct {
	repo repo.Repository
	log  *zap.SugaredLogger
}

// NewService - конструктор сервиса
func NewService(repo repo.Repository, logger *zap.SugaredLogger) Service {
	return &service{
		repo: repo,
		log:  logger,
	}
}

// CreateTask - обработчик запроса на создание задачи
func (s *service) CreateTask(ctx *fiber.Ctx) error {
	var req TaskRequest

	if err := ctx.BodyParser(&req); err != nil {
		s.log.Error("Invalid request body", zap.Error(err))
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	// Валидация входных данных
	if vErr := validator.Validate(ctx.Context(), req); vErr != nil {
		s.log.Warn("Validation failed", zap.String("error", vErr.Error()))
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid input data",
			"details": vErr.Error(),
		})
	}

	// Вставка задачи в БД через репозиторий
	task := repo.Task{
		Title:       req.Title,
		Description: req.Description,
	}
	taskID, err := s.repo.CreateTask(ctx.Context(), task)
	if err != nil {
		s.log.Error("Database error: failed to insert task", zap.Error(err))
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create task",
		})
	}
	s.log.Info("Task created successfully", zap.Int("task_id", taskID))
	// Формирование ответа
	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status":  "success",
		"task_id": taskID,
	})

}
func (s *service) GetTaskByID(ctx *fiber.Ctx) error {
	id, err := strconv.Atoi(ctx.Params("id"))
	if err != nil {
		s.log.Error("Invalid task ID format", zap.String("id", ctx.Params("id")), zap.Error(err))
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid task ID format",
			"details": "Task ID must be an integer",
		})
	}

	// Запрашиваем задачу в репозитории
	task, err := s.repo.GetTaskByID(ctx.Context(), id)
	if err != nil {
		s.log.Error("Database error: failed to retrieve task", zap.Int("task_id", id), zap.Error(err))
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve task",
		})
	}
	if task == nil {
		s.log.Warn("Task not found", zap.Int("task_id", id))
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Task not found",
		})
	}
	s.log.Info("Task retrieved successfully", zap.Int("task_id", id))

	// Возвращаем успешный ответ
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success",
		"data":   task,
	})
}

func (s *service) UpdateTask(ctx *fiber.Ctx) error {
	id, err := strconv.Atoi(ctx.Params("id"))
	if err != nil {
		s.log.Error("Invalid task ID format", zap.Error(err))
		return dto.BadResponseError(ctx, dto.FieldIncorrect, "Invalid task ID format")
	}

	var req TaskRequest
	// Десериализация JSON-запроса
	if err := ctx.BodyParser(&req); err != nil {
		s.log.Error("Invalid request body", zap.Error(err))
		return dto.BadResponseError(ctx, dto.FieldBadFormat, "Invalid request body")
	}

	// Валидация входных данных
	if vErr := validator.Validate(ctx.Context(), req); vErr != nil {
		return dto.BadResponseError(ctx, dto.FieldIncorrect, vErr.Error())
	}

	// Обновление задачи в БД через репозиторий
	task := repo.Task{
		Title:       req.Title,
		Description: req.Description,
	}
	err = s.repo.UpdateTask(ctx.Context(), id, task)
	if err != nil {
		s.log.Error("Failed to update task", zap.Error(err))
		return dto.InternalServerError(ctx)
	}

	// Формирование ответа
	response := dto.Response{
		Status: "success",
		Data:   map[string]int{"task_id": id},
	}

	return ctx.Status(fiber.StatusOK).JSON(response)
}
func (s *service) DeleteTask(ctx *fiber.Ctx) error {
	id, err := strconv.Atoi(ctx.Params("id"))
	if err != nil {
		s.log.Error("Invalid task ID format", zap.Error(err))
		return dto.BadResponseError(ctx, dto.FieldIncorrect, "Invalid task ID format")
	}
	existingTask, err := s.repo.GetTaskByID(ctx.Context(), id)
	if err != nil {
		s.log.Error("Database error while checking task existence", zap.Error(err))
		return dto.InternalServerError(ctx)
	}
	if existingTask == nil {
		s.log.Warn("Task not found", zap.Int("task_id", id))
		return dto.NotFoundError(ctx, "Task not found")
	}
	err = s.repo.DeleteTask(ctx.Context(), id)
	if err != nil {
		s.log.Error("Failed to delete task", zap.Error(err))
		return dto.InternalServerError(ctx)
	}

	// Формирование ответа
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": fmt.Sprintf("Task with ID %d has been deleted", id),
	})
}
func (s *service) GetTasksByUserID(ctx *fiber.Ctx) error {
	userID, err := strconv.Atoi(ctx.Params("userID"))
	if err != nil {
		s.log.Error("Invalid user ID format", zap.Error(err))
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID format",
		})
	}

	tasks, err := s.repo.GetTasksByUserID(ctx.Context(), userID)
	if err != nil {
		s.log.Error("Database error: failed to retrieve tasks", zap.Error(err))
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve tasks",
		})
	}

	if len(tasks) == 0 {
		return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
			"status": "success",
			"data":   []interface{}{},
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success",
		"data":   tasks,
	})
}
