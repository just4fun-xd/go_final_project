package service

import (
	"database/sql"
	"go_final_project/model"
)

// TaskRepository содержит методы для работы с хранилищем задач
type TaskRepository struct {
	DB *sql.DB
}

// NewTaskRepository создаёт новый экземпляр TaskRepository
func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{DB: db}
}

// CreateTask сохраняет задачу в базе данных
func (r *TaskRepository) CreateTask(task model.Tasks) (int64, error) {
	query := "INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)"
	result, err := r.DB.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// GetTaskByID получает задачу по идентификатору
func (r *TaskRepository) GetTaskByID(id int) (model.Tasks, error) {
	var task model.Tasks
	query := "SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?"
	err := r.DB.QueryRow(query, id).Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	return task, err
}

// UpdateTask обновляет задачу в базе данных
func (r *TaskRepository) UpdateTask(task model.Tasks) (int64, error) {
	query := "UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?"
	result, err := r.DB.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		return 0, err
	}
	affectedRows, err := result.RowsAffected()
	return affectedRows, err
}

// DeleteTask удаляет задачу по идентификатору
func (r *TaskRepository) DeleteTask(id int) (int64, error) {
	query := "DELETE FROM scheduler WHERE id = ?"
	result, err := r.DB.Exec(query, id)
	if err != nil {
		return 0, err
	}
	affectedRows, err := result.RowsAffected()
	return affectedRows, err
}

// GetAllTasks получает все задачи
func (r *TaskRepository) GetAllTasks(limit int) ([]model.Tasks, error) {
	rows, err := r.DB.Query("SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date LIMIT ?", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []model.Tasks
	for rows.Next() {
		var task model.Tasks
		if err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, rows.Err()
}

func (r *TaskRepository) MarkTaskAsDone(id int) (int64, error) {
	query := "UPDATE scheduler SET done = 1 WHERE id = ?"
	result, err := r.DB.Exec(query, id)
	if err != nil {
		return 0, err
	}
	affectedRows, err := result.RowsAffected()
	return affectedRows, err
}
