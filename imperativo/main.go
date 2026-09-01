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

type ErrorResponse struct {
	Error string `json:"error"`
}

func main() {
	// =========================
	// BANCO DE DADOS
	// =========================

	db, err := sql.Open("sqlite", "banco.db")
	if err != nil {
		log.Fatal("Erro ao abrir banco:", err)
	}

	defer db.Close()

	// SQLite funciona melhor com uma conexão de escrita por vez
	db.SetMaxOpenConns(1)

	err = db.Ping()
	if err != nil {
		log.Fatal("Erro ao conectar ao SQLite:", err)
	}

	// =========================
	// CRIAÇÃO DA TABELA
	// =========================

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS task (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			completed BOOLEAN NOT NULL DEFAULT FALSE
		)
	`)

	if err != nil {
		log.Fatal("Erro ao criar tabela:", err)
	}

	log.Println("Banco SQLite iniciado com sucesso")

	// =========================
	// ROTA PRINCIPAL
	// =========================

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w)

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Evita que qualquer rota inexistente caia aqui
		if r.URL.Path != "/" {
			writeError(
				w,
				"Rota não encontrada",
				http.StatusNotFound,
			)
			return
		}

		writeJSON(
			w,
			http.StatusOK,
			map[string]string{
				"message": "API Go funcionando",
			},
		)
	})

	// =========================
	// GET /todos
	// POST /todos
	// =========================

	http.HandleFunc("/todos", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w)

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		switch r.Method {

		// =====================
		// GET /todos
		// =====================

		case http.MethodGet:

			rows, err := db.Query(`
				SELECT id, title, completed
				FROM task
				ORDER BY id
			`)

			if err != nil {
				writeError(
					w,
					"Erro ao buscar tarefas",
					http.StatusInternalServerError,
				)
				return
			}

			defer rows.Close()

			// Importante:
			// [] em vez de null quando não existir tarefa
			tasks := []Task{}

			for rows.Next() {
				var task Task

				err := rows.Scan(
					&task.ID,
					&task.Title,
					&task.Completed,
				)

				if err != nil {
					writeError(
						w,
						"Erro ao ler tarefa",
						http.StatusInternalServerError,
					)
					return
				}

				tasks = append(tasks, task)
			}

			if err := rows.Err(); err != nil {
				writeError(
					w,
					"Erro ao percorrer tarefas",
					http.StatusInternalServerError,
				)
				return
			}

			writeJSON(
				w,
				http.StatusOK,
				tasks,
			)

		// =====================
		// POST /todos
		// =====================

		case http.MethodPost:

			var task Task

			err := json.NewDecoder(r.Body).Decode(&task)

			if err != nil {
				writeError(
					w,
					"JSON inválido",
					http.StatusBadRequest,
				)
				return
			}

			task.Title = strings.TrimSpace(task.Title)

			if task.Title == "" {
				writeError(
					w,
					"O título é obrigatório",
					http.StatusBadRequest,
				)
				return
			}

			// O frontend envia somente title.
			// Nova tarefa começa como não concluída.
			task.Completed = false

			result, err := db.Exec(`
				INSERT INTO task (
					title,
					completed
				)
				VALUES (?, ?)
			`,
				task.Title,
				task.Completed,
			)

			if err != nil {
				writeError(
					w,
					"Erro ao criar tarefa",
					http.StatusInternalServerError,
				)
				return
			}

			id, err := result.LastInsertId()

			if err != nil {
				writeError(
					w,
					"Erro ao obter ID da tarefa",
					http.StatusInternalServerError,
				)
				return
			}

			task.ID = int(id)

			writeJSON(
				w,
				http.StatusCreated,
				task,
			)

		default:

			writeError(
				w,
				"Método não permitido",
				http.StatusMethodNotAllowed,
			)
		}
	})

	// =========================
	// /todos/{id}
	// /todos/{id}/toggle
	// =========================

	http.HandleFunc("/todos/", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w)

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		path := strings.TrimPrefix(
			r.URL.Path,
			"/todos/",
		)

		parts := strings.Split(path, "/")

		if len(parts) == 0 || parts[0] == "" {
			writeError(
				w,
				"ID inválido",
				http.StatusBadRequest,
			)
			return
		}

		id, err := strconv.Atoi(parts[0])

		if err != nil || id <= 0 {
			writeError(
				w,
				"ID inválido",
				http.StatusBadRequest,
			)
			return
		}

		// =====================
		// PATCH /todos/{id}/toggle
		// =====================

		if r.Method == http.MethodPatch {

			if len(parts) != 2 || parts[1] != "toggle" {
				writeError(
					w,
					"Rota PATCH inválida",
					http.StatusNotFound,
				)
				return
			}

			// Alterna:
			// false -> true
			// true  -> false
			result, err := db.Exec(`
				UPDATE task
				SET completed = NOT completed
				WHERE id = ?
			`, id)

			if err != nil {
				writeError(
					w,
					"Erro ao atualizar tarefa",
					http.StatusInternalServerError,
				)
				return
			}

			rowsAffected, err := result.RowsAffected()

			if err != nil {
				writeError(
					w,
					"Erro ao verificar atualização",
					http.StatusInternalServerError,
				)
				return
			}

			if rowsAffected == 0 {
				writeError(
					w,
					"Tarefa não encontrada",
					http.StatusNotFound,
				)
				return
			}

			// Busca a tarefa já atualizada
			var task Task

			err = db.QueryRow(`
				SELECT id, title, completed
				FROM task
				WHERE id = ?
			`, id).Scan(
				&task.ID,
				&task.Title,
				&task.Completed,
			)

			if err != nil {
				writeError(
					w,
					"Erro ao buscar tarefa atualizada",
					http.StatusInternalServerError,
				)
				return
			}

			writeJSON(
				w,
				http.StatusOK,
				task,
			)

			return
		}

		// =====================
		// DELETE /todos/{id}
		// =====================

		if r.Method == http.MethodDelete {

			// DELETE aceita somente /todos/{id}
			if len(parts) != 1 {
				writeError(
					w,
					"Rota DELETE inválida",
					http.StatusNotFound,
				)
				return
			}

			result, err := db.Exec(
				"DELETE FROM task WHERE id = ?",
				id,
			)

			if err != nil {
				writeError(
					w,
					"Erro ao deletar tarefa",
					http.StatusInternalServerError,
				)
				return
			}

			rowsAffected, err := result.RowsAffected()

			if err != nil {
				writeError(
					w,
					"Erro ao verificar exclusão",
					http.StatusInternalServerError,
				)
				return
			}

			if rowsAffected == 0 {
				writeError(
					w,
					"Tarefa não encontrada",
					http.StatusNotFound,
				)
				return
			}

			w.WriteHeader(http.StatusNoContent)

			return
		}

		writeError(
			w,
			"Método não permitido",
			http.StatusMethodNotAllowed,
		)
	})

	// =========================
	// SERVIDOR
	// =========================

	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	log.Println("Servidor rodando na porta " + port)

	log.Fatal(
		http.ListenAndServe(
			"0.0.0.0:"+port,
			nil,
		),
	)
}

// =========================
// CORS
// =========================

func enableCORS(w http.ResponseWriter) {
	w.Header().Set(
		"Access-Control-Allow-Origin",
		"*",
	)

	w.Header().Set(
		"Access-Control-Allow-Methods",
		"GET, POST, PATCH, DELETE, OPTIONS",
	)

	w.Header().Set(
		"Access-Control-Allow-Headers",
		"Content-Type, Accept",
	)
}

// =========================
// RESPOSTA JSON
// =========================

func writeJSON(
	w http.ResponseWriter,
	status int,
	data interface{},
) {
	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(status)

	err := json.NewEncoder(w).Encode(data)

	if err != nil {
		log.Println(
			"Erro ao gerar JSON:",
			err,
		)
	}
}

// =========================
// RESPOSTA DE ERRO
// =========================

func writeError(
	w http.ResponseWriter,
	message string,
	status int,
) {
	writeJSON(
		w,
		status,
		ErrorResponse{
			Error: message,
		},
	)
}
