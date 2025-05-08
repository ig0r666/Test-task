package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"testtask/limiter/adapters/rest/middleware"
	"testtask/limiter/core"
)

func MainHandler(rate core.RateLimiter, db core.RateLimiterDB) http.HandlerFunc {
	handler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Allowed =)")
	}

	return middleware.Rate(handler, rate, db)
	// Передаем хендлер в лимитер. Если у клиента
	// остались токены, то пропускаем его дальше
}

// CRUD Для работы с клиентами

// CreateClientHandler - POST /clients
// Создает клиента с заданным client_id и capacity
// Принимает JSON вида:
// {"client_id": "string", "capacity": int}
func CreateClientHandler(log *slog.Logger, db core.CrudDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req core.ClientRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Error("failed to decode request", "error", err)
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if req.ClientID == "" || req.Capacity <= 0 {
			http.Error(w, "client_id and capacity are required", http.StatusBadRequest)
			return
		}

		client := core.Client{
			ClientID: req.ClientID,
			Capacity: req.Capacity,
			Tokens:   req.Capacity, // Создаем клиента с полным набором токенов по умолчанию
		}

		if err := db.CreateClient(r.Context(), client); err != nil {
			log.Error("failed to create client", "error", err)
			http.Error(w, "failed to create client", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"status": "client created"})
	}
}

// GetClientsHandler - GET /clients
// Выводит список всех клиентов с заданным capacity в формате JSON
func GetClientsHandler(log *slog.Logger, db core.CrudDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clients, err := db.GetAllClients(r.Context())
		if err != nil {
			log.Error("failed to get clients", "error", err)
			http.Error(w, "failed to get clients", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(clients)
	}
}

// GetClientHandler - GET /client?client_id={id}
// Возвращает клиента с заданным client_id в формате JSON
func GetClientHandler(log *slog.Logger, db core.CrudDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientID := r.URL.Query().Get("client_id")
		if clientID == "" {
			http.Error(w, "client_id is required", http.StatusBadRequest)
			return
		}

		clientDb, err := db.GetClient(r.Context(), clientID)
		if err != nil {
			log.Error("failed to get client", "error", err)
			http.Error(w, "failed to get client", http.StatusInternalServerError)
			return
		}

		client := core.ClientRequest{
			ClientID: clientDb.ClientID,
			Capacity: clientDb.Capacity,
			Tokens:   clientDb.Tokens,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(client)
	}
}

// DeleteClientHandler - DELETE /client?client_id={id}
// Удаляет клиента с заданным client_id
func DeleteClientHandler(log *slog.Logger, db core.CrudDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientID := r.URL.Query().Get("client_id")
		if clientID == "" {
			http.Error(w, "client_id is required", http.StatusBadRequest)
			return
		}

		if err := db.RemoveClient(r.Context(), clientID); err != nil {
			log.Error("failed to delete client", "client_id", clientID, "error", err)

			if errors.Is(err, core.ErrClientNotFound) {
				http.Error(w, "client_id not found", http.StatusBadRequest)
				return
			}
			http.Error(w, "failed to delete client", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "client successful deleted"})
	}
}

// UpdateClientHandler - PUT /client
// Обновляет capacity заданного клиента
// Принимает JSON вида:
// {"client_id": "string", "capacity": int}
func UpdateClientHandler(log *slog.Logger, db core.CrudDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req core.ClientRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Error("failed to decode request", "error", err)
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if req.ClientID == "" || req.Capacity <= 0 {
			http.Error(w, "client_id and capacity are required", http.StatusBadRequest)
			return
		}

		if err := db.UpdateClientCapacity(r.Context(), req.ClientID, req.Capacity); err != nil {
			log.Error("failed to update client", "error", err)
			http.Error(w, "failed to update client", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "client successful updated"})
	}
}
