package service

import (
	"time"
)

type TaskService struct{}

func NewTaskService() *TaskService {
	return &TaskService{}
}

func (s *TaskService) NextDate(now time.Time, dateStr string, repeat string) (string, error) {
	return NextDate(now, dateStr, repeat)
}
