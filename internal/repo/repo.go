package repo

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"

	"simple-service/internal/config"
)

// Слой репозитория, здесь должны быть все методы, связанные с базой данных

// SQL-запрос на вставку задачи
const (
	insertTaskQuery     = `INSERT INTO tasks (title, description) VALUES ($1, $2) RETURNING id;`
	selectTaskByID      = `SELECT id, user_id, title, description, status, created_at FROM tasks WHERE id = $1`
	updateTaskQuery     = `UPDATE tasks SET title = $1, description = $2 WHERE id = $3`
	deleteTaskQuery     = `DELETE FROM tasks WHERE id = $1`
	selectTasksByUserID = `SELECT id, user_id, title, description, status, created_at FROM tasks WHERE user_id = $1`
)

type repository struct {
	pool *pgxpool.Pool
}

// Repository - интерфейс с методом создания задачи
type Repository interface {
	CreateTask(ctx context.Context, task Task) (int, error)
	GetTaskByID(ctx context.Context, id int) (*Task, error)
	DeleteTask(ctx context.Context, id int) error
	UpdateTask(ctx context.Context, id int, task Task) error
	GetTasksByUserID(ctx context.Context, userID int) ([]Task, error)
}

// NewRepository - создание нового экземпляра репозитория с подключением к PostgreSQL
func NewRepository(ctx context.Context, cfg config.PostgreSQL) (Repository, error) {
	// Формируем строку подключения
	connString := fmt.Sprintf(
		`user=%s password=%s host=%s port=%d dbname=%s sslmode=%s 
        pool_max_conns=%d pool_max_conn_lifetime=%s pool_max_conn_idle_time=%s`,
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
		cfg.SSLMode,
		cfg.PoolMaxConns,
		cfg.PoolMaxConnLifetime.String(),
		cfg.PoolMaxConnIdleTime.String(),
	)

	// Парсим конфигурацию подключения
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse PostgreSQL config")
	}

	// Оптимизация выполнения запросов (кеширование запросов)
	config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeCacheDescribe

	// Создаём пул соединений с базой данных
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create PostgreSQL connection pool")
	}

	return &repository{pool}, nil
}

// CreateTask - вставка новой задачи в таблицу tasks
func (r *repository) CreateTask(ctx context.Context, task Task) (int, error) {
	var id int
	err := r.pool.QueryRow(ctx, insertTaskQuery, task.Title, task.Description).Scan(&id)
	if err != nil {
		return 0, errors.Wrap(err, "failed to insert task 1")
	}
	return id, nil
}
func (r *repository) GetTaskByID(ctx context.Context, id int) (*Task, error) {
	var task Task
	err := r.pool.QueryRow(ctx, selectTaskByID, id).Scan(&task.ID, &task.UserID, &task.Title, &task.Description, &task.Status, &task.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // Возвращаем nil, если задача не найдена
		}
		return nil, errors.Wrap(err, "failed to retrieve task")
	}
	return &task, nil
}

// UpdateTask - обновление задачи
func (r *repository) UpdateTask(ctx context.Context, id int, task Task) error {
	result, err := r.pool.Exec(ctx, updateTaskQuery, task.Title, task.Description, id)
	if err != nil {
		return errors.Wrap(err, "failed to update task")
	}
	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows // Если строк не изменилось, возвращаем ошибку "нет данных"
	}
	return nil
}

// DeleteTask - удаление задачи
func (r *repository) DeleteTask(ctx context.Context, id int) error {
	result, err := r.pool.Exec(ctx, deleteTaskQuery, id)
	if err != nil {
		return errors.Wrap(err, "failed to delete task")
	}
	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows // Если строк не удалено, значит задачи не было
	}
	return nil

}

// GetTasksByUserID - получение задач по ID пользователя
func (r *repository) GetTasksByUserID(ctx context.Context, userID int) ([]Task, error) {
	rows, err := r.pool.Query(ctx, selectTasksByUserID, userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query tasks by user_id")
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		if err := rows.Scan(&task.ID, &task.UserID, &task.Title, &task.Description, &task.Status, &task.CreatedAt); err != nil {
			return nil, errors.Wrap(err, "failed to scan task by user_id")
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "error while iterating over user tasks")
	}

	return tasks, nil
}
