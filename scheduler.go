package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

func NextDate(now time.Time, dateStr string, repeat string) (string, error) {
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
		days, err := strconv.Atoi(rule[1])
		if err != nil || days <= 0 || days > 400 {
			return "", fmt.Errorf("наверное количество дней: %v", rule[1])
		}
		nextDate := taskDate.AddDate(0, 0, days)
		for nextDate.Before(now) {
			nextDate = nextDate.AddDate(0, 0, days)
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
}
