package service

import (
	"time"
)

// TaskService содержит методы для работы с задачами
type TaskService struct{}

// NewTaskService создаёт новый экземпляр TaskService
func NewTaskService() *TaskService {
	return &TaskService{}
}

// NextDate вычисляет следующую дату
func (s *TaskService) NextDate(now time.Time, dateStr string, repeat string) (string, error) {
	return NextDate(now, dateStr, repeat)
}
