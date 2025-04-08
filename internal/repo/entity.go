package repo

import "time"

// Task - структура, соответствующая таблице tasks
type Task struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}
type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
}
