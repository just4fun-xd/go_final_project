package service

import (
	"database/sql"
	"go_final_project/model"
)

type TaskRepository struct {
	DB *sql.DB
}

func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{DB: db}
}

func (r *TaskRepository) CreateTask(task model.Tasks) (int64, error) {
	query := "INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)"
	result, err := r.DB.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *TaskRepository) GetTaskByID(id int) (model.Tasks, error) {
	var task model.Tasks
	query := "SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?"
	err := r.DB.QueryRow(query, id).Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	return task, err
}

func (r *TaskRepository) UpdateTask(task model.Tasks) (int64, error) {
	query := "UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?"
	result, err := r.DB.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		return 0, err
	}
	affectedRows, err := result.RowsAffected()
	return affectedRows, err
}

func (r *TaskRepository) UpdateTaskDate(id int, nextDate string) (int64, error) {
	query := "UPDATE scheduler SET date = ? WHERE id = ?"
	result, err := r.DB.Exec(query, nextDate, id)
	if err != nil {
		return 0, err
	}
	affectedRows, err := result.RowsAffected()
	return affectedRows, err
}

func (r *TaskRepository) DeleteTask(id int) (int64, error) {
	query := "DELETE FROM scheduler WHERE id = ?"
	result, err := r.DB.Exec(query, id)
	if err != nil {
		return 0, err
	}
	affectedRows, err := result.RowsAffected()
	return affectedRows, err
}

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
