package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
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

	// Rota /todos
	http.HandleFunc("/todos", func(w http.ResponseWriter, r *http.Request) {

		// =========================
		// GET /todos
		// =========================
		if r.Method == http.MethodGet {

			rows, err := db.Query(`
				SELECT id, title, completed
				FROM task
			`)

			if err != nil {
				http.Error(w, "Erro ao buscar tarefas", 500)
				return
			}

			defer rows.Close()

			var tasks []Task

			for rows.Next() {

				var task Task

				err := rows.Scan(
					&task.ID,
					&task.Title,
					&task.Completed,
				)

				if err != nil {
					http.Error(w, "Erro ao ler tarefa", 500)
					return
				}

				tasks = append(tasks, task)
			}

			w.Header().Set("Content-Type", "application/json")

			json.NewEncoder(w).Encode(tasks)

			return
		}

		// =========================
		// POST /todos
		// =========================
		if r.Method == http.MethodPost {

			var task Task

			err := json.NewDecoder(r.Body).Decode(&task)

			if err != nil {
				http.Error(w, "JSON inválido", 400)
				return
			}

			result, err := db.Exec(`
				INSERT INTO task (title, completed)
				VALUES (?, ?)
			`,
				task.Title,
				task.Completed,
			)

			if err != nil {
				http.Error(w, "Erro ao criar tarefa", 500)
				return
			}

			id, err := result.LastInsertId()

			if err != nil {
				http.Error(w, "Erro ao obter ID", 500)
				return
			}

			task.ID = int(id)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)

			json.NewEncoder(w).Encode(task)

			return
		}

		// Método não permitido
		http.Error(
			w,
			"Método não permitido",
			http.StatusMethodNotAllowed,
		)
	})

	// Rota DELETE /todos/{id}
	http.HandleFunc("/todos/", func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodDelete {
			http.Error(
				w,
				"Método não permitido",
				http.StatusMethodNotAllowed,
			)
			return
		}

		// Exemplo:
		// /todos/10
		path := strings.TrimPrefix(r.URL.Path, "/todos/")

		id, err := strconv.Atoi(path)

		if err != nil {
			http.Error(w, "ID inválido", 400)
			return
		}

		result, err := db.Exec(
			"DELETE FROM task WHERE id = ?",
			id,
		)

		if err != nil {
			http.Error(w, "Erro ao deletar tarefa", 500)
			return
		}

		rowsAffected, err := result.RowsAffected()

		if err != nil {
			http.Error(w, "Erro ao verificar exclusão", 500)
			return
		}

		if rowsAffected == 0 {
			http.Error(w, "Tarefa não encontrada", 404)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})

	log.Println("Servidor rodando em http://localhost:8080")

	err = http.ListenAndServe(":8080", nil)

	if err != nil {
		log.Fatal(err)
	}
}