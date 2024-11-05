package api

import (
	"database/sql"
	"encoding/json"
	"go_final_project/model"
	"go_final_project/service"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"
)

type Handlers struct {
	TaskService    *service.TaskService
	TaskRepository *service.TaskRepository
}

func NewHandlers(db *sql.DB) *Handlers {
	return &Handlers{
		TaskService:    service.NewTaskService(),
		TaskRepository: service.NewTaskRepository(db),
	}
}

func writeErrorResponse(w http.ResponseWriter, statusCode int, errMsg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := model.TaskResponse{Error: errMsg}
	json.NewEncoder(w).Encode(response)
}

func (h *Handlers) PostTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task model.Tasks
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
	task.Date, err = service.ValidateTaskDate(now, task.Date, task.Repeat)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	taskID, err := h.TaskRepository.CreateTask(task)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Ошибка базы данных: "+err.Error())
		return
	}

	response := model.TaskResponse{ID: strconv.Itoa(int(taskID))}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handlers) GetTaskHandler(w http.ResponseWriter, r *http.Request) {
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

	task, err := h.TaskRepository.GetTaskByID(id)
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

func (h *Handlers) GetNextDateHandler(w http.ResponseWriter, r *http.Request) {
	nowStr := r.URL.Query().Get("now")
	dateStr := r.URL.Query().Get("date")
	repeat := r.URL.Query().Get("repeat")

	if nowStr == "" || dateStr == "" || repeat == "" {
		http.Error(w, "Отсутствует требуемый параметр", http.StatusBadRequest)
		return
	}

	now, err := time.Parse(service.DateFormat, nowStr)
	if err != nil {
		log.Printf("Ошибка: не удалось разобрать дату 'now': %v", err)
		http.Error(w, "Неверный формат 'now'", http.StatusBadRequest)
		return
	}

	nextDate, err := h.TaskService.NextDate(now, dateStr, repeat)
	if err != nil {
		log.Printf("Ошибка при вычислении следующей даты: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(nextDate))
}

func (h *Handlers) GetTasksHandler(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.TaskRepository.GetAllTasks(service.TaskQueryLimit)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Ошибка выполнения запроса: "+err.Error())
		return
	}

	if tasks == nil {
		tasks = []model.Tasks{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"tasks": tasks,
	})
}

func (h *Handlers) PutTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task model.Tasks
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Ошибка декодирования JSON: "+err.Error())
		return
	}
	if task.ID == "0" {
		writeErrorResponse(w, http.StatusBadRequest, "Не указан идентификатор задачи")
		return
	}
	if task.Title == "" {
		writeErrorResponse(w, http.StatusBadRequest, "Не указан заголовок задачи")
		return
	}

	id, err := strconv.Atoi(task.ID)
	if err != nil || id <= 0 || id > math.MaxInt32 {
		writeErrorResponse(w, http.StatusBadRequest, "Некорректный идентификатор задачи")
		return
	}

	now := time.Now()
	task.Date, err = service.ValidateTaskDate(now, task.Date, task.Repeat)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	_, err = h.TaskRepository.UpdateTask(task)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Ошибка базы данных: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("{}")); err != nil {
		log.Printf("Ошибка записи ответа: %v", err)
	}
}

func (h *Handlers) DeleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		writeErrorResponse(w, http.StatusBadRequest, "Не указан идентификатор задачи")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Некорректный идентификатор задачи")
		return
	}

	_, err = h.TaskRepository.DeleteTask(id)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Ошибка базы данных: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("{}")); err != nil {
		log.Printf("Ошибка записи ответа: %v", err)
	}
}

func (h *Handlers) DoneTaskHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		writeErrorResponse(w, http.StatusBadRequest, "Не указан идентификатор задачи.")
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Некорректный идентификатор задачи.")
		return
	}

	task, err := h.TaskRepository.GetTaskByID(id)
	if err == sql.ErrNoRows {
		writeErrorResponse(w, http.StatusNotFound, "Задача не найдена.")
		return
	} else if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Ошибка выполнения запроса: "+err.Error())
		return
	}

	if task.Repeat == "" {
		_, err = h.TaskRepository.DeleteTask(id)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "Ошибка удаления задачи: "+err.Error())
			return
		}
	} else {
		now := time.Now()
		nextDate, err := h.TaskService.NextDate(now, task.Date, task.Repeat)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "Ошибка при расчете следующей даты: "+err.Error())
			return
		}

		_, err = h.TaskRepository.UpdateTaskDate(id, nextDate)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "Ошибка обновления задачи: "+err.Error())
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("{}")); err != nil {
		log.Printf("Ошибка записи ответа: %v", err)
	}
}
