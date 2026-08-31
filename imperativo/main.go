package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	_ "modernc.org/sqlite"
)

type Task struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

func main() {

	// Abre o banco
	db, err := sql.Open("sqlite", "banco.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Cria a tabela
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS task (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			completed BOOLEAN NOT NULL
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Rota /todos -> GET (listar) e POST (criar)
	http.HandleFunc("/todos", func(w http.ResponseWriter, r *http.Request) {

		switch r.Method {

		case http.MethodGet:
			rows, err := db.Query(`SELECT id, title, completed FROM task`)
			if err != nil {
				http.Error(w, "Erro ao buscar tarefas", http.StatusInternalServerError)
				return
			}
			defer rows.Close()

			var tasks []Task
			for rows.Next() {
				var task Task
				if err := rows.Scan(&task.ID, &task.Title, &task.Completed); err != nil {
					http.Error(w, "Erro ao ler tarefa", http.StatusInternalServerError)
					return
				}
				tasks = append(tasks, task)
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(tasks)

		case http.MethodPost:
			var task Task
			if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
				http.Error(w, "JSON inválido", http.StatusBadRequest)
				return
			}

			result, err := db.Exec(
				"INSERT INTO task (title, completed) VALUES (?, ?)",
				task.Title, task.Completed,
			)
			if err != nil {
				http.Error(w, "Erro ao criar tarefa", http.StatusInternalServerError)
				return
			}

			id, err := result.LastInsertId()
			if err != nil {
				http.Error(w, "Erro ao obter ID", http.StatusInternalServerError)
				return
			}
			task.ID = int(id)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(task)

		default:
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		}
	})

	// Rota /todos/{id} -> PATCH (completar) e DELETE (remover)
	http.HandleFunc("/todos/", func(w http.ResponseWriter, r *http.Request) {

		path := strings.TrimPrefix(r.URL.Path, "/todos/")
		id, err := strconv.Atoi(path)
		if err != nil {
			http.Error(w, "ID inválido", http.StatusBadRequest)
			return
		}

		switch r.Method {

		case http.MethodPatch:
			result, err := db.Exec("UPDATE task SET completed = ? WHERE id = ?", true, id)
			if err != nil {
				http.Error(w, "Erro ao atualizar tarefa", http.StatusInternalServerError)
				return
			}
			rowsAffected, _ := result.RowsAffected()
			if rowsAffected == 0 {
				http.Error(w, "Tarefa não encontrada", http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusOK)

		case http.MethodDelete:
			result, err := db.Exec("DELETE FROM task WHERE id = ?", id)
			if err != nil {
				http.Error(w, "Erro ao deletar tarefa", http.StatusInternalServerError)
				return
			}
			rowsAffected, _ := result.RowsAffected()
			if rowsAffected == 0 {
				http.Error(w, "Tarefa não encontrada", http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		}
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Servidor rodando na porta " + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}