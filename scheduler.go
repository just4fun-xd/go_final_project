package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

func NextDate(now time.Time, dateStr string, repeat string) (string, error) {
	if repeat == "" {
		return "", fmt.Errorf("правило повторения не указано")
	}
	taskDate, err := time.Parse("20060102", dateStr)
	if err != nil {
		return "", fmt.Errorf("неверная дата: %v", err)
	}

	if repeat == "d 1" && !taskDate.After(now) {
		return now.Format("20060102"), nil
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
			return taskDate.Format("20060102"), nil
		}
	}
}

func NextDateOld(now time.Time, dateStr string, repeat string) (string, error) {
	log.Printf("Вычисляем следующую дату для исходной даты: %s с правилом: %s", dateStr, repeat)

	taskDate, err := time.Parse("20060102", dateStr)
	if err != nil {
		log.Printf("Ошибка: не удалось разобрать исходную дату: %v", err)
		return "", fmt.Errorf("неверная дата: %v", err)
	}
	if repeat == "" {
		log.Printf("Ошибка: правило повторения не указано")
		return "", fmt.Errorf("правило повторения не указано")
	}

	rule := strings.Split(repeat, " ")
	if len(rule) == 0 {
		return "", fmt.Errorf("неверный формат правила повторения")
	}

	if rule[0] == "d" {
		if len(rule) != 2 {
			return "", fmt.Errorf("неверный формат правила d")
		}

		if rule[1] == "" {
			return "", fmt.Errorf("не указано количество дней в правиле d")
		}

		days, err := strconv.Atoi(rule[1])

		if err != nil || days <= 0 || days > 400 {
			return "", fmt.Errorf("наверное количество дней: %v", rule[1])
		}

		nextDate := taskDate.AddDate(0, 0, days)
		for nextDate.Before(now) {
			nextDate = nextDate.AddDate(0, 0, days)
		}
		if days == 1 && now.Format("20060102") == taskDate.Format("20060102") {
			log.Printf("Задача уже сегодня: %s", taskDate.Format("20060102"))
			return taskDate.Format("20060102"), nil
		}

		log.Printf("Следующая дата вычислена: %s", nextDate.Format("20060102"))
		return nextDate.Format("20060102"), nil
	}

	if rule[0] == "y" {
		if len(rule) != 1 {
			return "", fmt.Errorf("неверный формат правила y")
		}
		nextDate := taskDate.AddDate(1, 0, 0)
		for nextDate.Before(now) {
			nextDate = nextDate.AddDate(1, 0, 0)
		}
		log.Printf("Следующая дата вычислена: %s", nextDate.Format("20060102"))
		return nextDate.Format("20060102"), nil
	}

	return "", fmt.Errorf("неподдерживаемое правило повторения: %s", repeat)
}

func NextDateTest(now time.Time, dateStr string, repeat string) (string, error) {
	if repeat == "" {
		return "", fmt.Errorf("правило повторения не указано")
	}
	fmt.Printf("dateStr = %s\n", dateStr)

	var taskDate time.Time
	var err error

	// Если dateStr пустое, используем текущую дату
	if dateStr == "" {
		taskDate = now
	} else {
		taskDate, err = time.Parse("20060102", dateStr)
		if err != nil {
			return "", fmt.Errorf("неверная дата: %v", err)
		}
	}

	if strings.HasPrefix(repeat, "d ") && (dateStr == "") {
		// Устанавливаем задачу на сегодняшнюю дату без применения правила повторения
		return taskDate.Format("20060102"), nil
	}

	// Применяем правило повторения
	for {
		if repeat == "y" {
			taskDate = taskDate.AddDate(1, 0, 0)
		} else if strings.HasPrefix(repeat, "d ") {
			daysStr := strings.TrimPrefix(repeat, "d ")
			days, err := strconv.Atoi(daysStr)
			if err != nil || days < 1 || days > 400 {
				return "", fmt.Errorf("неверное количество дней: %v", days)
			}
			taskDate = taskDate.AddDate(0, 0, days)
		} else {
			return "", fmt.Errorf("неподдерживаемое правило повторения: %s", repeat)
		}

		// Если задача теперь находится в будущем, возвращаем её
		if taskDate.After(now) {
			return taskDate.Format("20060102"), nil
		}
	}
}
