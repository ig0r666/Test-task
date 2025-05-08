package db

import (
	"context"
	"log/slog"
	"testtask/limiter/core"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

type DB struct {
	log  *slog.Logger
	conn *sqlx.DB
}

func New(log *slog.Logger, address string) (*DB, error) {

	db, err := sqlx.Connect("pgx", address)
	if err != nil {
		log.Error("connection problem", "address", address, "error", err)
		return nil, err
	}

	return &DB{
		log:  log,
		conn: db,
	}, nil
}

func (db *DB) GetClient(ctx context.Context, clientID string) (core.Client, error) {
	const query = `
		SELECT client_id, capacity, tokens FROM client WHERE client_id = $1 FOR UPDATE;
	`

	var client core.Client
	err := db.conn.GetContext(ctx, &client, query, clientID)
	if err != nil {
		db.log.Error("failed to get client", "client_id", clientID, "error", err)
		return core.Client{}, core.ErrClientNotFound
	}

	return client, nil
}

func (db *DB) GetAllClients(ctx context.Context) ([]core.Client, error) {
	const query = `
        SELECT client_id, capacity, tokens FROM client;
    `

	var clients []core.Client
	err := db.conn.SelectContext(ctx, &clients, query)
	if err != nil {
		db.log.Error("failed to get clients", "error", err)
		return nil, err
	}

	return clients, nil
}

func (db *DB) UpdateClientToken(ctx context.Context, clientID string, token int) error {
	const query = `
		UPDATE client 
		SET tokens = $1
		WHERE client_id = $2;
	`

	result, err := db.conn.ExecContext(ctx, query, token, clientID)
	if err != nil {
		db.log.Error("failed to update tokens", "client_id", clientID, "error", err)
		return err
	}

	rowsChanged, _ := result.RowsAffected()
	if rowsChanged == 0 {
		db.log.Warn("client not found", "client_id", clientID)
		return core.ErrClientNotFound
	}

	return nil
}

func (db *DB) UpdateClientCapacity(ctx context.Context, clientID string, capacity int) error {
	const query = `
		UPDATE client 
		SET capacity = $1
		WHERE client_id = $2;
	`

	result, err := db.conn.ExecContext(ctx, query, capacity, clientID)
	if err != nil {
		db.log.Error("failed to update capacity", "client_id", clientID, "error", err)
		return err
	}

	rowsChanged, _ := result.RowsAffected()
	if rowsChanged == 0 {
		db.log.Warn("client not found", "client_id", clientID)
		return core.ErrClientNotFound
	}

	return nil
}

func (db *DB) CreateClient(ctx context.Context, client core.Client) error {
	const query = `
		INSERT INTO client (client_id, capacity, tokens)
		VALUES (:client_id, :capacity, :tokens)
		ON CONFLICT (client_id) 
		DO NOTHING;
	`

	_, err := db.conn.NamedExecContext(ctx, query, client)
	if err != nil {
		db.log.Error("failed to create client",
			"client_id", client.ClientID,
			"error", err)
		return err
	}

	return nil
}

func (db *DB) RemoveClient(ctx context.Context, clientID string) error {
	const query = `
		DELETE FROM client 
		WHERE client_id = $1
		RETURNING client_id;
	`

	var deletedID string
	err := db.conn.QueryRowxContext(ctx, query, clientID).Scan(&deletedID)
	if err != nil {
		db.log.Error("failed to delete client", "client_id", clientID, "error", err)
		if err.Error() == "sql: no rows in result set" {
			return core.ErrClientNotFound
		}
		return err
	}

	return nil
}

func (db *DB) UpdateAllTokens(ctx context.Context) error {
	// Оборачиваем операции в транзакцию для потокобезопасности
	tx, err := db.conn.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	const query = `
		UPDATE client
		SET tokens = capacity
		WHERE tokens < capacity;
	`

	_, err = tx.ExecContext(ctx, query)
	if err != nil {
		db.log.Error("failed to update tokens", "error", err)
		return err
	}

	db.log.Debug("tokens updated for all clients")
	return tx.Commit()
}
