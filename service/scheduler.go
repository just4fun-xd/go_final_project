package service

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func NextDate(now time.Time, dateStr string, repeat string) (string, error) {
	if repeat == "" {
		return "", fmt.Errorf("правило повторения не указано")
	}
	taskDate, err := time.Parse(DateFormat, dateStr)
	if err != nil {
		return "", fmt.Errorf("неверная дата: %v", err)
	}

	if repeat == "d 1" && !taskDate.After(now) {
		return now.Format(DateFormat), nil
	}

	for {
		if repeat == "y" {
			taskDate = taskDate.AddDate(1, 0, 0)
		} else if strings.HasPrefix(repeat, "d ") {
			daysStr := strings.TrimPrefix(repeat, "d ")
			days, err := strconv.Atoi(daysStr)
			if err != nil || days < 1 || days > 400 {
				return "", fmt.Errorf("наверное количество дней: %v", days)
			}
			taskDate = taskDate.AddDate(0, 0, days)
		} else {
			return "", fmt.Errorf("неподдерживаемое правило повторения: %s", repeat)
		}
		if taskDate.After(now) {
			return taskDate.Format(DateFormat), nil
		}
	}
}
