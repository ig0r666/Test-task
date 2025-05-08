package main

import (
	"flag"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"testtask/balancer/adapters/roundrobin"
	"testtask/balancer/adapters/server"
	"testtask/balancer/config"
	"testtask/balancer/core"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "config.yaml", "server configuration file")
	flag.Parse()

	cfg := config.MustLoad(configPath)

	log := mustMakeLogger(cfg.LogLevel)
	log.Info("starting server")

	// Достаем серверы из конфига
	servers := createServers(cfg.ServersURLs, log)
	pool := roundrobin.NewRoundRobin(log, nil)
	lb := core.NewLoadBalancer(log, pool)

	// Инициализируем балансировщик
	if err := lb.Initialize(servers); err != nil {
		log.Error("failed to init servers", "error", err)
		os.Exit(1)
	}

	// В горутине запускаем healthcheck для проверки серверов с заданным интервалом
	go lb.StartHealthCheck(cfg.HealthCheckInterval)

	// Поднимаем сервер
	mux := http.NewServeMux()
	mux.HandleFunc("/", lb.Handler)

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

func createServers(serverURLs string, log *slog.Logger) []core.Server {
	var servers []core.Server

	urls := strings.Split(serverURLs, ",")
	for _, u := range urls {
		u = strings.TrimSpace(u)
		if u == "" {
			continue
		}

		serverURL, err := url.Parse(u)
		if err != nil {
			log.Error("Failed to parse server URL", "url", u, "error", err)
			continue
		}

		proxy := httputil.NewSingleHostReverseProxy(serverURL)
		servers = append(servers, server.NewServer(serverURL, proxy))
	}

	return servers
}
