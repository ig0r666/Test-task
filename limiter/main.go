package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"testtask/limiter/adapters/db"
	"testtask/limiter/adapters/ratelimiter"
	"testtask/limiter/adapters/rest"
	"testtask/limiter/config"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "config.yaml", "server configuration file")
	flag.Parse()

	// Создаем конфиг
	cfg := config.MustLoad(configPath)

	// Создаем логгер
	log := mustMakeLogger(cfg.LogLevel)
	log.Info("starting server")

	// Инициализируем бд, проводим миграции
	storage, err := db.New(log, cfg.DBAddress)
	if err != nil {
		log.Error("failed to connect to db", "error", err)
		os.Exit(1)
	}
	if err := storage.Migrate(); err != nil {
		log.Error("failed to migrate db", "error", err)
		os.Exit(1)
	}

	// Инициализируем лимитер
	ctx, cancel := context.WithCancel(context.Background())
	rl := ratelimiter.New(ctx, log, cfg, storage)
	defer cancel()

	// Добавляем обработчики для эндпоинтов
	mux := http.NewServeMux()
	mux.HandleFunc("GET /test", rest.MainHandler(rl, storage))

	mux.HandleFunc("POST /clients", rest.CreateClientHandler(log, storage))
	mux.HandleFunc("GET /clients", rest.GetClientsHandler(log, storage))
	mux.HandleFunc("GET /client", rest.GetClientHandler(log, storage))
	mux.HandleFunc("DELETE /client", rest.DeleteClientHandler(log, storage))
	mux.HandleFunc("PUT /client", rest.UpdateClientHandler(log, storage))

	server := http.Server{
		Addr:        cfg.HTTPConfig.Address,
		Handler:     mux,
		ReadTimeout: cfg.HTTPConfig.Timeout,
	}

	log.Info("starting server", "address", cfg.HTTPConfig.Address)
	if err := server.ListenAndServe(); err != nil {
		log.Error("server error", "error", err)
	}
}

func mustMakeLogger(logLevel string) *slog.Logger {
	var level slog.Level
	switch logLevel {
	case "DEBUG":
		level = slog.LevelDebug
	case "INFO":
		level = slog.LevelInfo
	case "ERROR":
		level = slog.LevelError
	default:
		panic("unknown log level: " + logLevel)
	}
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	return slog.New(handler)
}
