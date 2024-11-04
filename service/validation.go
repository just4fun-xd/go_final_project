package service

import (
	"fmt"
	"time"
)

// ValidateTaskDate проверяет дату задачи и возвращает её
func ValidateTaskDate(now time.Time, taskDateStr string, repeat string) (string, error) {
	if taskDateStr == "" {
		return now.Format(DateFormat), nil
	}

	taskDate, err := time.Parse(DateFormat, taskDateStr)
	if err != nil {
		return "", fmt.Errorf("некорректная дата. Ожидается формат 20060102: %v", err)
	}

	parsedDate := taskDate.Truncate(24 * time.Hour)
	now = now.Truncate(24 * time.Hour)

	if parsedDate.Before(now) {
		if repeat == "" {
			return now.Format(DateFormat), nil
		}
		nextDate, err := NextDate(now, taskDateStr, repeat)
		if err != nil {
			return "", fmt.Errorf("ошибка в правиле повторения: %v", err)
		}
		return nextDate, nil
	}

	return taskDateStr, nil
}
