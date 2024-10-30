package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Tasks struct {
	ID      string `json:"id,omitempty"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

type TaskResponse struct {
	ID    string `json:"id,omitempty"`
	Error string `json:"error,omitempty"`
}

func writeErrorResponse(w http.ResponseWriter, statusCode int, errMsr string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := TaskResponse{Error: errMsr}
	json.NewEncoder(w).Encode(response)
}

func PostTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task Tasks
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Ошибка декодирования JSON: "+err.Error())
		return
	}

	if task.Title == "" {
		writeErrorResponse(w, http.StatusBadRequest, "Не указан заголовок задачи")
		return
	}

	now := time.Now()
	today := now.Format("20060102")

	if task.Date == "" {
		task.Date = today
	} else {
		taskDate, err := time.Parse("20060102", task.Date)
		if err != nil {
			writeErrorResponse(w, http.StatusBadRequest, "Некорректная дата. Ожидается формат 20060102")
			return
		}

		parsedDate := taskDate.Truncate(24 * time.Hour)
		now = now.Truncate(24 * time.Hour)

		if parsedDate.Before(now) {
			if task.Repeat == "" {
				task.Date = today
			} else {
				nextDate, err := NextDate(now, task.Date, task.Repeat)
				if err != nil {
					writeErrorResponse(w, http.StatusBadRequest, "Ошибка в правиле повторения: "+err.Error())
					return
				}
				if task.Repeat == "d 1" && nextDate == today {
					task.Date = today
				} else {
					task.Date = nextDate
				}
			}
		}
	}

	if task.Repeat != "" {
		_, err := NextDate(now, task.Date, task.Repeat)
		if err != nil {
			writeErrorResponse(w, http.StatusBadRequest, "Некорректное правило повторения: "+err.Error())
			return
		}
	}

	db, err := sql.Open("sqlite3", "./scheduler.db")
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Ошибка подключения к базе данных: "+err.Error())
		return
	}
	defer db.Close()

	query := "INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)"
	result, err := db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Ошибка базы данных: "+err.Error())
		return
	}
	taskID, err := result.LastInsertId()
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Ошибка при получении id задачи:"+err.Error())
		return
	}

	response := TaskResponse{ID: strconv.Itoa(int(taskID))}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func GetTaskHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		writeErrorResponse(w, http.StatusBadRequest, "Не указан идентификатор")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Некорректный идентификатор")
		return
	}

	db, err := sql.Open("sqlite3", "./scheduler.db")
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Ошибка подключения к базе данных: "+err.Error())
		return
	}
	defer db.Close()

	var task Tasks
	query := "SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?"
	err = db.QueryRow(query, id).Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err == sql.ErrNoRows {
		writeErrorResponse(w, http.StatusNotFound, "Задача не найдена")
		return
	} else if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Ошибка выполнения запроса: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(task)
}

func GetNextDateHandler(w http.ResponseWriter, r *http.Request) {
	nowStr := r.URL.Query().Get("now")
	dateStr := r.URL.Query().Get("date")
	repeat := r.URL.Query().Get("repeat")

	if nowStr == "" || dateStr == "" || repeat == "" {
		http.Error(w, "Отсутствует требуемый параметр", http.StatusBadRequest)
		return
	}

	now, err := time.Parse("20060102", nowStr)
	if err != nil {
		log.Printf("Ошибка: не удалось разобрать дату 'now': %v", err)
		http.Error(w, "Неверный формат 'now'", http.StatusBadRequest)
		return
	}

	nextDate, err := NextDate(now, dateStr, repeat)
	if err != nil {
		log.Printf("Ошибка при вычислении следующей даты: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(nextDate))
}

func GetTasksHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("sqlite3", "./scheduler.db")
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Ошибка подключения к базе данных: "+err.Error())
		return
	}
	defer db.Close()

	now := time.Now().Format("20060102")

	rows, err := db.Query("SELECT id, date, title, comment, repeat FROM scheduler WHERE date >= ? ORDER BY date LIMIT 50", now)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Ошибка выполнения запроса: "+err.Error())
		return
	}
	defer rows.Close()

	var tasks []Tasks
	for rows.Next() {
		var task Tasks

		if err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "Ошибка при чтении результата: "+err.Error())
			return
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Ошибка обработки результата: "+err.Error())
		return
	}

	if tasks == nil {
		tasks = []Tasks{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"tasks": tasks,
	})
}

func PutTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task Tasks
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Ошибка декодирования JSON: "+err.Error())
		return
	}
	fmt.Println(task)
	if task.ID == "0" {
		writeErrorResponse(w, http.StatusBadRequest, "Не указан идентификатор задачи")
		return
	}
	if task.Title == "" {
		writeErrorResponse(w, http.StatusBadRequest, "Не указан заголовок задачи")
		return
	}

	now := time.Now()
	if task.Date == "" {
		task.Date = now.Format("20060102")
	}

	taskDate, err := time.Parse("20060102", task.Date)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Некорректная дата. Ожидается формат 20060102")
		return
	}

	if taskDate.Before(now) {
		if task.Repeat == "" {
			task.Date = now.Format("20060102")
		} else {
			newDate, err := NextDate(now, task.Date, task.Repeat)
			if err != nil {
				writeErrorResponse(w, http.StatusBadRequest, "Некорректное правило повторения: "+err.Error())
				return
			}
			task.Date = newDate
		}
	} else if task.Repeat != "" {
		newDate, err := NextDate(now, task.Date, task.Repeat)
		if err != nil {
			writeErrorResponse(w, http.StatusBadRequest, "Некорректное правило повторения: "+err.Error())
			return
		}
		task.Date = newDate
	}

	db, err := sql.Open("sqlite3", "./scheduler.db")
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Ошибка подключения к базе данных: "+err.Error())
		return
	}
	defer db.Close()

	// Обновляем обработчик PUT
	query := "UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?"
	result, err := db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Ошибка базы данных: "+err.Error())
		return
	}

	rowsAffected, err := result.RowsAffected()
	log.Printf("Обновлено строк: %d\n", rowsAffected) // <-- добавьте это для проверки

	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Ошибка при получении количества обновленных строк: "+err.Error())
		return
	}

	if rowsAffected == 0 {
		writeErrorResponse(w, http.StatusNotFound, "Задача не найдена")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
}

func DoneTaskHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		writeErrorResponse(w, http.StatusBadRequest, "Не указан идентификатор задачи.")
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Некоректный идентификатор задачи.")
	}

	db, err := sql.Open("sqlite3", "./scheduler.db")
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Ошибка подключения базы данных: "+err.Error())
		return
	}

	var task Tasks
	query := "SELECT id, date, repeat FROM scheduler WHERE id = ?"
	err = db.QueryRow(query, id).Scan(&task.ID, &task.Date, &task.Repeat)
	if err == sql.ErrNoRows {
		writeErrorResponse(w, http.StatusNotFound, "Задача не найдена.")
		return
	} else if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Ошибка выполнения запроса: "+err.Error())
		return
	}
	if task.Repeat == "" {
		_, err = db.Exec("DELETE FROM scheduler WHERE id = ?", id)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "Ошибка удаления задачи: "+err.Error())
			return
		}
	} else {
		now := time.Now()
		nextDate, err := NextDate(now, task.Date, task.Repeat)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "Ошибка при расчете следующей даты: "+err.Error())
			return
		}

		_, err = db.Exec("UPDATE scheduler SET date = ? WHERE id = ?", nextDate, id)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "Ошибка обновления задачи: "+err.Error())
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
}

func DeleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		writeErrorResponse(w, http.StatusBadRequest, "Не указан идентификатор задачи.")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Некоректный идентификатор задачи.")
		return
	}

	db, err := sql.Open("sqlite3", "./scheduler.db")
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Ошибка подключения базы данных: "+err.Error())
		return
	}
	defer db.Close()

	query := "DELETE FROM scheduler WHERE id = ?"
	result, err := db.Exec(query, id)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Ошибка базы данных: "+err.Error())
		return
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Ошибка при получении количества удаленных строк: "+err.Error())
		return
	}

	if rowsAffected == 0 {
		writeErrorResponse(w, http.StatusNotFound, "Задача не найдена")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
}
