package main

import (
	"database/sql"
	"go_final_project/tests"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

// InitDB инициализирует базу данных
func InitDB() *sql.DB {
	dbFile := os.Getenv("TODO_DBFILE")

	if dbFile == "" {
		appPath, err := os.Executable()
		if err != nil {
			log.Fatal("Ошибка получения пути к файлу:", err)
		}
		dbFile = filepath.Join(filepath.Dir(appPath), "scheduler.db")
	}

	_, err := os.Stat(dbFile)

	var install bool
	if err != nil {
		install = true
	}

	db, err := sql.Open("sqlite3", "./scheduler.db")
	if err != nil {
		log.Fatal("Ошибка при открытии базы данных:", err)
	}

	if install {
		_, err = db.Exec(`
			CREATE TABLE IF NOT EXISTS scheduler (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				date CHAR(8) NOT NULL DEFAULT "",
				title VARCHAR(256) NOT NULL DEFAULT "",
				comment TEXT,
				repeat VARCHAR(128) NOT NULL DEFAULT ""
			);
			CREATE INDEX IF NOT EXISTS scheduler_date ON scheduler (date);
		`)
		if err != nil {
			log.Fatal("Ошибка при создании таблицы:", err)
		}
		log.Println("База данных создана успешно.")

	}
	return db
}

func main() {
	webDir := "./web"

	db := InitDB()
	defer db.Close()

	fileServer := http.FileServer(http.Dir(webDir))
	http.Handle("/", fileServer)

	http.HandleFunc("/api/nextdate", GetNextDateHandler)
	http.HandleFunc("/api/tasks", GetTasksHandler)
	http.HandleFunc("/api/task/done", DoneTaskHandler)
	http.HandleFunc("/api/task", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			GetTaskHandler(w, r)
		case http.MethodPost:
			PostTaskHandler(w, r)
		case http.MethodPut:
			PutTaskHandler(w, r)
		case http.MethodDelete:
			DeleteTaskHandler(w, r)
		default:
			http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		}
	})

	defaultPort := tests.Port
	envPort := os.Getenv("TODO_PORT")

	port := defaultPort
	if envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			port = p
		} else {
			log.Printf("Некорректное значение переменной TODO_PORT: %s. Использую порт по умолчанию: %d", envPort, defaultPort)
		}
	}

	log.Printf("Сервер запущен на порте %d\n", port)

	err := http.ListenAndServe(":"+strconv.Itoa(port), nil)
	if err != nil {
		log.Fatal("Ошибка запуска сервера:", err)
	}
}
